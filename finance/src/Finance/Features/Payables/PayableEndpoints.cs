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

        if (!await db.Stakeholders.AnyAsync(s => s.Id == supplierId && s.Type == StakeholderType.Supplier, cancellationToken))
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
