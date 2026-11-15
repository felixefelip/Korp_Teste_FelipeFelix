using Finance.Data;
using Finance.Domain;
using Finance.Features.Spreadsheets;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.EntityFrameworkCore;

namespace Finance.Features.Stakeholders;

public static class StakeholderImportEndpoints
{
    public static void MapStakeholderImport(this IEndpointRouteBuilder app)
    {
        app.MapPost("/stakeholders/import", Import)
            .WithTags("Stakeholders")
            .DisableAntiforgery();
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
        catch (DbUpdateException exception) when (exception.IsUniqueViolation())
        {
            return TypedResults.Problem(
                title: "Stakeholder already exists.",
                detail: "A document in the spreadsheet was registered by another request. Try again.",
                statusCode: StatusCodes.Status409Conflict);
        }

        return TypedResults.Ok(stakeholders.Select(StakeholderEndpoints.ToDto).ToArray());
    }
}
