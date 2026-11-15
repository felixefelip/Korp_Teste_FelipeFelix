using Finance.Data;
using Finance.Domain;
using Finance.Features.Spreadsheets;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.Data.SqlClient;
using Microsoft.EntityFrameworkCore;

namespace Finance.Features.Stakeholders;

public static class StakeholderEndpoints
{
    private const int UniqueIndexViolation = 2601;
    private const int UniqueConstraintViolation = 2627;

    public static void MapStakeholders(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/stakeholders").WithTags("Stakeholders");

        group.MapPost("/", Create);
        group.MapGet("/{id:int}", GetById);
        group.MapPost("/import", Import).DisableAntiforgery();
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
        catch (DbUpdateException exception) when (IsDuplicate(exception))
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

    private static async Task<Results<Ok<StakeholderDto[]>, ValidationProblem, ProblemHttpResult>> Import(
        IFormFile file, StakeholderType? type, FinanceDbContext db, CancellationToken cancellationToken)
    {
        if (type is not { } stakeholderType || !Enum.IsDefined(stakeholderType))
        {
            return TypedResults.ValidationProblem(new Dictionary<string, string[]>
            {
                ["Type"] = ["Type is required and must be Customer or Supplier."]
            });
        }

        var spreadsheet = RegistrationSpreadsheet.Read(file);

        if (spreadsheet.Errors.Count > 0)
        {
            return TypedResults.ValidationProblem(spreadsheet.Errors);
        }

        var documents = spreadsheet.Rows.Select(row => row.Document).Distinct().ToArray();
        var registeredDocuments = new HashSet<string>();

        foreach (var chunk in documents.Chunk(1000))
        {
            registeredDocuments.UnionWith(await db.Stakeholders
                .Where(s => s.Type == stakeholderType && chunk.Contains(s.Document))
                .Select(s => s.Document)
                .ToListAsync(cancellationToken));
        }

        var errors = RegistrationSpreadsheet.Validate(spreadsheet.Rows, registeredDocuments);

        if (errors.Count > 0)
        {
            return TypedResults.ValidationProblem(errors);
        }

        var stakeholders = spreadsheet.Rows
            .Select(row => new Stakeholder(row.Name, row.Document, stakeholderType))
            .ToList();
        db.Stakeholders.AddRange(stakeholders);

        try
        {
            await db.SaveChangesAsync(cancellationToken);
        }
        catch (DbUpdateException exception) when (IsDuplicate(exception))
        {
            return TypedResults.Problem(
                title: "Stakeholder already exists.",
                detail: "A document in the spreadsheet was registered by another request. Try again.",
                statusCode: StatusCodes.Status409Conflict);
        }

        return TypedResults.Ok(stakeholders.Select(ToDto).ToArray());
    }

    private static bool IsDuplicate(DbUpdateException exception) =>
        exception.InnerException is SqlException { Number: UniqueIndexViolation or UniqueConstraintViolation };

    private static StakeholderDto ToDto(Stakeholder stakeholder) => new(
        stakeholder.Id,
        stakeholder.Name,
        stakeholder.Document,
        stakeholder.Type,
        stakeholder.CreatedAt);
}
