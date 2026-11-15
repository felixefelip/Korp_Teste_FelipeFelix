using Finance.Data;
using Finance.Domain;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.EntityFrameworkCore;

namespace Finance.Features.Receivables;

public static class ReceivablePaymentEndpoints
{
    public static void MapReceivablePayments(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/receivables/{id:int}/payments").WithTags("Receivables");

        group.MapPost("/", Register);
        group.MapGet("/", List);
        group.MapGet("/{paymentId:int}", GetById);
    }

    private static async Task<Results<Created<ReceivablePaymentDto>, NotFound, ValidationProblem, ProblemHttpResult>> Register(
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

    private static async Task<Results<Ok<ReceivablePaymentDto[]>, NotFound>> List(
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

    private static async Task<Results<Ok<ReceivablePaymentDto>, NotFound>> GetById(
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
}
