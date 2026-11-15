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
