using Finance.Data;
using Finance.Domain;
using Finance.Features.Spreadsheets;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.Data.SqlClient;
using Microsoft.EntityFrameworkCore;

namespace Finance.Features.Customers;

public static class CustomerEndpoints
{
    private const int UniqueIndexViolation = 2601;
    private const int UniqueConstraintViolation = 2627;

    public static void MapCustomers(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/customers").WithTags("Customers");

        group.MapPost("/", Create);
        group.MapGet("/{id:int}", GetById);
        group.MapPost("/import", Import).DisableAntiforgery();
    }

    private static async Task<Results<Created<CustomerDto>, ProblemHttpResult>> Create(
        CreateCustomerDto dto, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var customer = new Customer(dto.Name, dto.Document);

        db.Customers.Add(customer);

        try
        {
            await db.SaveChangesAsync(cancellationToken);
        }
        catch (DbUpdateException exception) when (IsDuplicate(exception))
        {
            return TypedResults.Problem(
                title: "Customer already exists.",
                detail: "A customer with this document is already registered.",
                statusCode: StatusCodes.Status409Conflict);
        }

        return TypedResults.Created($"/customers/{customer.Id}", ToDto(customer));
    }

    private static async Task<Results<Ok<CustomerDto>, NotFound>> GetById(
        int id, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var customer = await db.Customers.FindAsync([id], cancellationToken);

        return customer is null
            ? TypedResults.NotFound()
            : TypedResults.Ok(ToDto(customer));
    }

    private static async Task<Results<Ok<CustomerDto[]>, ValidationProblem, ProblemHttpResult>> Import(
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
            registeredDocuments.UnionWith(await db.Customers
                .Where(c => chunk.Contains(c.Document))
                .Select(c => c.Document)
                .ToListAsync(cancellationToken));
        }

        var errors = RegistrationSpreadsheet.Validate(spreadsheet.Rows, registeredDocuments);

        if (errors.Count > 0)
        {
            return TypedResults.ValidationProblem(errors);
        }

        var customers = spreadsheet.Rows.Select(row => new Customer(row.Name, row.Document)).ToList();
        db.Customers.AddRange(customers);

        try
        {
            await db.SaveChangesAsync(cancellationToken);
        }
        catch (DbUpdateException exception) when (IsDuplicate(exception))
        {
            return TypedResults.Problem(
                title: "Customer already exists.",
                detail: "A document in the spreadsheet was registered by another request. Try again.",
                statusCode: StatusCodes.Status409Conflict);
        }

        return TypedResults.Ok(customers.Select(ToDto).ToArray());
    }

    private static bool IsDuplicate(DbUpdateException exception) =>
        exception.InnerException is SqlException { Number: UniqueIndexViolation or UniqueConstraintViolation };

    private static CustomerDto ToDto(Customer customer) => new(
        customer.Id,
        customer.Name,
        customer.Document,
        customer.CreatedAt);
}
