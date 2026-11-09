using Finance.Data;
using Finance.Domain;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.EntityFrameworkCore;

namespace Finance.Features.Payables;

public static class PayableEndpoints
{
    public static void MapPayables(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/payables").WithTags("Payables");

        group.MapPost("/", Create);
        group.MapGet("/{id:int}", GetById);
        group.MapPost("/{id:int}/payments", RegisterPayment);
        group.MapGet("/{id:int}/payments", ListPayments);
        group.MapGet("/{id:int}/payments/{paymentId:int}", GetPaymentById);
    }

    private static async Task<Results<Created<PayableDto>, ValidationProblem>> Create(
        CreatePayableDto dto, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var bankAccountId = dto.BankAccountId!.Value;
        var supplierId = dto.SupplierId!.Value;
        var errors = new Dictionary<string, string[]>();

        if (!await db.BankAccounts.AnyAsync(b => b.Id == bankAccountId, cancellationToken))
        {
            errors[nameof(dto.BankAccountId)] = ["Bank account not found."];
        }

        if (!await db.Suppliers.AnyAsync(c => c.Id == supplierId, cancellationToken))
        {
            errors[nameof(dto.SupplierId)] = ["Supplier not found."];
        }

        if (errors.Count > 0)
        {
            return TypedResults.ValidationProblem(errors);
        }

        var payable = new Payable(
            bankAccountId, dto.Description, supplierId, dto.Amount!.Value, dto.DueDate!.Value);

        db.Payables.Add(payable);
        await db.SaveChangesAsync(cancellationToken);

        return TypedResults.Created($"/payables/{payable.Id}", ToDto(payable));
    }

    private static async Task<Results<Ok<PayableDto>, NotFound>> GetById(
        int id, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var payable = await db.Payables.FindAsync([id], cancellationToken);

        return payable is null
            ? TypedResults.NotFound()
            : TypedResults.Ok(ToDto(payable));
    }

    private static async Task<Results<Created<PayablePaymentDto>, NotFound, ValidationProblem, ProblemHttpResult>> RegisterPayment(
        int id, RegisterPayablePaymentDto dto, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var payable = await db.Payables.FindAsync([id], cancellationToken);

        if (payable is null)
        {
            return TypedResults.NotFound();
        }

        if (payable.Status == EntryStatus.Paid)
        {
            return TypedResults.Problem(
                title: "Payable already paid.",
                detail: "There is nothing left to pay on this payable.",
                statusCode: StatusCodes.Status409Conflict);
        }

        var amount = dto.Amount!.Value;

        if (amount > payable.RemainingAmount)
        {
            return TypedResults.ValidationProblem(new Dictionary<string, string[]>
            {
                [nameof(dto.Amount)] = ["Amount exceeds the remaining amount."]
            });
        }

        var payment = payable.RegisterPayment(amount, dto.PaidOn!.Value);

        try
        {
            await db.SaveChangesAsync(cancellationToken);
        }
        catch (DbUpdateConcurrencyException)
        {
            return TypedResults.Problem(
                title: "Payable changed by another request.",
                detail: "Reload the payable and try again.",
                statusCode: StatusCodes.Status409Conflict);
        }

        return TypedResults.Created($"/payables/{id}/payments/{payment.Id}", ToDto(payment));
    }

    private static async Task<Results<Ok<PayablePaymentDto[]>, NotFound>> ListPayments(
        int id, FinanceDbContext db, CancellationToken cancellationToken)
    {
        if (!await db.Payables.AnyAsync(p => p.Id == id, cancellationToken))
        {
            return TypedResults.NotFound();
        }

        var payments = await db.PayablePayments
            .Where(p => p.PayableId == id)
            .OrderBy(p => p.PaidOn)
            .ThenBy(p => p.Id)
            .ToArrayAsync(cancellationToken);

        return TypedResults.Ok(payments.Select(ToDto).ToArray());
    }

    private static async Task<Results<Ok<PayablePaymentDto>, NotFound>> GetPaymentById(
        int id, int paymentId, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var payment = await db.PayablePayments
            .FirstOrDefaultAsync(p => p.Id == paymentId && p.PayableId == id, cancellationToken);

        return payment is null
            ? TypedResults.NotFound()
            : TypedResults.Ok(ToDto(payment));
    }

    private static PayablePaymentDto ToDto(PayablePayment payment) => new(
        payment.Id,
        payment.PayableId,
        payment.Amount,
        payment.PaidOn,
        payment.CreatedAt);

    private static PayableDto ToDto(Payable payable) => new(
        payable.Id,
        payable.BankAccountId,
        payable.Description,
        payable.SupplierId,
        payable.Amount,
        payable.PaidAmount,
        payable.RemainingAmount,
        payable.DueDate,
        payable.Status,
        payable.CreatedAt);
}
