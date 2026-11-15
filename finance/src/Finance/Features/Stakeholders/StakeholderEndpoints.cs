using Finance.Data;
using Finance.Domain;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.EntityFrameworkCore;

namespace Finance.Features.Stakeholders;

public static class StakeholderEndpoints
{
    public static void MapStakeholders(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/stakeholders").WithTags("Stakeholders");

        group.MapPost("/", Create);
        group.MapGet("/{id:int}", GetById);
    }

    private static async Task<Results<Created<StakeholderDto>, ProblemHttpResult>> Create(
        CreateStakeholderDto dto, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var stakeholder = new Stakeholder(dto.Name, dto.Document, dto.Type!.Value);

        db.Stakeholders.Add(stakeholder);

        try
        {
            await db.SaveChangesAsync(cancellationToken);
        }
        catch (DbUpdateException exception) when (exception.IsUniqueViolation())
        {
            return TypedResults.Problem(
                title: "Stakeholder already exists.",
                detail: $"A {stakeholder.Type} with this document is already registered.",
                statusCode: StatusCodes.Status409Conflict);
        }

        return TypedResults.Created($"/stakeholders/{stakeholder.Id}", ToDto(stakeholder));
    }

    private static async Task<Results<Ok<StakeholderDto>, NotFound>> GetById(
        int id, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var stakeholder = await db.Stakeholders.FindAsync([id], cancellationToken);

        return stakeholder is null
            ? TypedResults.NotFound()
            : TypedResults.Ok(ToDto(stakeholder));
    }

    internal static StakeholderDto ToDto(Stakeholder stakeholder) => new(
        stakeholder.Id,
        stakeholder.Name,
        stakeholder.Document,
        stakeholder.Type,
        stakeholder.CreatedAt);
}
