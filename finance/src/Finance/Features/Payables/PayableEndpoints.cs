using Finance.Data;
using Finance.Domain;
using Microsoft.AspNetCore.Http.HttpResults;

namespace Finance.Features.Payables;

public static class PayableEndpoints
{
    public static void MapPayables(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/payables").WithTags("Payables");

        group.MapPost("/", Create);
        group.MapGet("/{id:int}", GetById);
    }

    private static async Task<Created<PayableDto>> Create(
        CreatePayableDto dto, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var payable = new Payable(
            dto.Description, dto.SupplierName, dto.Amount!.Value, dto.DueDate!.Value);

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

    private static PayableDto ToDto(Payable payable) => new(
        payable.Id,
        payable.Description,
        payable.SupplierName,
        payable.Amount,
        payable.PaidAmount,
        payable.RemainingAmount,
        payable.DueDate,
        payable.Status,
        payable.CreatedAt);
}
