using Finance.Data;
using Finance.Domain;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.EntityFrameworkCore;

namespace Finance.Features.Receivables;

public static class ReceivableEndpoints
{
    public static void MapReceivables(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/receivables").WithTags("Receivables");

        group.MapPost("/", Create);
        group.MapGet("/{id:int}", GetById);
        group.MapPost("/{id:int}/payments", RegisterPayment);
        group.MapGet("/{id:int}/payments", ListPayments);
        group.MapGet("/{id:int}/payments/{paymentId:int}", GetPaymentById);
    }

    private static async Task<Results<Created<ReceivableDto>, ValidationProblem>> Create(
        CreateReceivableDto dto, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var bankAccountId = dto.BankAccountId!.Value;
        var customerId = dto.CustomerId!.Value;
        var errors = new Dictionary<string, string[]>();

        if (!await db.BankAccounts.AnyAsync(b => b.Id == bankAccountId, cancellationToken))
        {
            errors[nameof(dto.BankAccountId)] = ["Bank account not found."];
        }

        if (!await db.Customers.AnyAsync(c => c.Id == customerId, cancellationToken))
        {
            errors[nameof(dto.CustomerId)] = ["Customer not found."];
        }

        if (errors.Count > 0)
        {
            return TypedResults.ValidationProblem(errors);
        }

        var receivable = new Receivable(
            bankAccountId, dto.Description, customerId, dto.Amount!.Value, dto.DueDate!.Value);

        db.Receivables.Add(receivable);
        await db.SaveChangesAsync(cancellationToken);

        return TypedResults.Created($"/receivables/{receivable.Id}", ToDto(receivable));
    }

    private static async Task<Results<Ok<ReceivableDto>, NotFound>> GetById(
        int id, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var receivable = await db.Receivables.FindAsync([id], cancellationToken);

        return receivable is null
            ? TypedResults.NotFound()
            : TypedResults.Ok(ToDto(receivable));
    }

    private static async Task<Results<Created<ReceivablePaymentDto>, NotFound, ValidationProblem, ProblemHttpResult>> RegisterPayment(
        int id, RegisterReceivablePaymentDto dto, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var receivable = await db.Receivables.FindAsync([id], cancellationToken);

        if (receivable is null)
        {
            return TypedResults.NotFound();
        }

        if (receivable.Status == EntryStatus.Paid)
        {
            return TypedResults.Problem(
                title: "Receivable already paid.",
                detail: "There is nothing left to receive on this receivable.",
                statusCode: StatusCodes.Status409Conflict);
        }

        var amount = dto.Amount!.Value;

        if (amount > receivable.RemainingAmount)
        {
            return TypedResults.ValidationProblem(new Dictionary<string, string[]>
            {
                [nameof(dto.Amount)] = ["Amount exceeds the remaining amount."]
            });
        }

        var payment = receivable.RegisterPayment(amount, dto.PaidOn!.Value);

        try
        {
            await db.SaveChangesAsync(cancellationToken);
        }
        catch (DbUpdateConcurrencyException)
        {
            return TypedResults.Problem(
                title: "Receivable changed by another request.",
                detail: "Reload the receivable and try again.",
                statusCode: StatusCodes.Status409Conflict);
        }

        return TypedResults.Created($"/receivables/{id}/payments/{payment.Id}", ToDto(payment));
    }

    private static async Task<Results<Ok<ReceivablePaymentDto[]>, NotFound>> ListPayments(
        int id, FinanceDbContext db, CancellationToken cancellationToken)
    {
        if (!await db.Receivables.AnyAsync(r => r.Id == id, cancellationToken))
        {
            return TypedResults.NotFound();
        }

        var payments = await db.ReceivablePayments
            .Where(p => p.ReceivableId == id)
            .OrderBy(p => p.PaidOn)
            .ThenBy(p => p.Id)
            .ToArrayAsync(cancellationToken);

        return TypedResults.Ok(payments.Select(ToDto).ToArray());
    }

    private static async Task<Results<Ok<ReceivablePaymentDto>, NotFound>> GetPaymentById(
        int id, int paymentId, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var payment = await db.ReceivablePayments
            .FirstOrDefaultAsync(p => p.Id == paymentId && p.ReceivableId == id, cancellationToken);

        return payment is null
            ? TypedResults.NotFound()
            : TypedResults.Ok(ToDto(payment));
    }

    private static ReceivablePaymentDto ToDto(ReceivablePayment payment) => new(
        payment.Id,
        payment.ReceivableId,
        payment.Amount,
        payment.PaidOn,
        payment.CreatedAt);

    private static ReceivableDto ToDto(Receivable receivable) => new(
        receivable.Id,
        receivable.BankAccountId,
        receivable.Description,
        receivable.CustomerId,
        receivable.Amount,
        receivable.PaidAmount,
        receivable.RemainingAmount,
        receivable.DueDate,
        receivable.Status,
        receivable.CreatedAt);
}
