using Finance.Data;
using Finance.Domain;
using Microsoft.AspNetCore.Http.HttpResults;

namespace Finance.Features.Receivables;

public static class ReceivableEndpoints
{
    public static void MapReceivables(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/receivables").WithTags("Receivables");

        group.MapPost("/", Create);
        group.MapGet("/{id:int}", GetById);
    }

    private static async Task<Created<ReceivableDto>> Create(
        CreateReceivableDto dto, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var receivable = new Receivable(
            dto.Description, dto.CustomerName, dto.Amount!.Value, dto.DueDate!.Value);

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
        receivable.Description,
        receivable.CustomerName,
        receivable.Amount,
        receivable.PaidAmount,
        receivable.RemainingAmount,
        receivable.DueDate,
        receivable.Status,
        receivable.CreatedAt);
}
