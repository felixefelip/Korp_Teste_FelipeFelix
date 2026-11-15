using Finance.Data;
using Finance.Domain;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.EntityFrameworkCore;

namespace Finance.Features.Payables;

public static class PayablePaymentEndpoints
{
    public static void MapPayablePayments(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/payables/{id:int}/payments").WithTags("Payables");

        group.MapPost("/", Register);
        group.MapGet("/", List);
        group.MapGet("/{paymentId:int}", GetById);
    }

    private static async Task<Results<Created<PayablePaymentDto>, NotFound, ValidationProblem, ProblemHttpResult>> Register(
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

    private static async Task<Results<Ok<PayablePaymentDto[]>, NotFound>> List(
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

    private static async Task<Results<Ok<PayablePaymentDto>, NotFound>> GetById(
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
}
