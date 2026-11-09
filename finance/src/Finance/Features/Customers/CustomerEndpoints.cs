using Finance.Data;
using Finance.Domain;
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

    private static bool IsDuplicate(DbUpdateException exception) =>
        exception.InnerException is SqlException { Number: UniqueIndexViolation or UniqueConstraintViolation };

    private static CustomerDto ToDto(Customer customer) => new(
        customer.Id,
        customer.Name,
        customer.Document,
        customer.CreatedAt);
}
