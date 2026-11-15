using Finance.Data;
using Finance.Domain;
using Finance.Features.Spreadsheets;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.Data.SqlClient;
using Microsoft.EntityFrameworkCore;

namespace Finance.Features.Suppliers;

public static class SupplierEndpoints
{
    private const int UniqueIndexViolation = 2601;
    private const int UniqueConstraintViolation = 2627;

    public static void MapSuppliers(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/suppliers").WithTags("Suppliers");

        group.MapPost("/", Create);
        group.MapGet("/{id:int}", GetById);
        group.MapPost("/import", Import).DisableAntiforgery();
    }

    private static async Task<Results<Created<SupplierDto>, ProblemHttpResult>> Create(
        CreateSupplierDto dto, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var supplier = new Supplier(dto.Name, dto.Document);

        db.Suppliers.Add(supplier);

        try
        {
            await db.SaveChangesAsync(cancellationToken);
        }
        catch (DbUpdateException exception) when (IsDuplicate(exception))
        {
            return TypedResults.Problem(
                title: "Supplier already exists.",
                detail: "A supplier with this document is already registered.",
                statusCode: StatusCodes.Status409Conflict);
        }

        return TypedResults.Created($"/suppliers/{supplier.Id}", ToDto(supplier));
    }

    private static async Task<Results<Ok<SupplierDto>, NotFound>> GetById(
        int id, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var supplier = await db.Suppliers.FindAsync([id], cancellationToken);

        return supplier is null
            ? TypedResults.NotFound()
            : TypedResults.Ok(ToDto(supplier));
    }

    private static async Task<Results<Ok<SupplierDto[]>, ValidationProblem, ProblemHttpResult>> Import(
        IFormFile file, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var spreadsheet = RegistrationSpreadsheet.Read(file);

        if (spreadsheet.Errors.Count > 0)
        {
            return TypedResults.ValidationProblem(spreadsheet.Errors);
        }

        var documents = spreadsheet.Rows.Select(row => row.Document).Distinct().ToArray();
        var registeredDocuments = new HashSet<string>();

        foreach (var chunk in documents.Chunk(1000))
        {
            registeredDocuments.UnionWith(await db.Suppliers
                .Where(s => chunk.Contains(s.Document))
                .Select(s => s.Document)
                .ToListAsync(cancellationToken));
        }

        var errors = RegistrationSpreadsheet.Validate(spreadsheet.Rows, registeredDocuments);

        if (errors.Count > 0)
        {
            return TypedResults.ValidationProblem(errors);
        }

        var suppliers = spreadsheet.Rows.Select(row => new Supplier(row.Name, row.Document)).ToList();
        db.Suppliers.AddRange(suppliers);

        try
        {
            await db.SaveChangesAsync(cancellationToken);
        }
        catch (DbUpdateException exception) when (IsDuplicate(exception))
        {
            return TypedResults.Problem(
                title: "Supplier already exists.",
                detail: "A document in the spreadsheet was registered by another request. Try again.",
                statusCode: StatusCodes.Status409Conflict);
        }

        return TypedResults.Ok(suppliers.Select(ToDto).ToArray());
    }

    private static bool IsDuplicate(DbUpdateException exception) =>
        exception.InnerException is SqlException { Number: UniqueIndexViolation or UniqueConstraintViolation };

    private static SupplierDto ToDto(Supplier supplier) => new(
        supplier.Id,
        supplier.Name,
        supplier.Document,
        supplier.CreatedAt);
}
