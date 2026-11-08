using Finance.Data;
using Finance.Domain;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.Data.SqlClient;
using Microsoft.EntityFrameworkCore;

namespace Finance.Features.BankAccounts;

public static class BankAccountEndpoints
{
    private const int UniqueIndexViolation = 2601;
    private const int UniqueConstraintViolation = 2627;

    public static void MapBankAccounts(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/bank-accounts").WithTags("BankAccounts");

        group.MapPost("/", Create);
        group.MapGet("/{id:int}", GetById);
    }

    private static async Task<Results<Created<BankAccountDto>, ProblemHttpResult>> Create(
        CreateBankAccountDto dto, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var bankAccount = new BankAccount(
            dto.Name, dto.BankCode, dto.Branch, dto.AccountNumber, dto.OpeningBalance!.Value);

        db.BankAccounts.Add(bankAccount);

        try
        {
            await db.SaveChangesAsync(cancellationToken);
        }
        catch (DbUpdateException exception) when (IsDuplicate(exception))
        {
            return TypedResults.Problem(
                title: "Bank account already exists.",
                detail: "A bank account with this bank code, branch and account number is already registered.",
                statusCode: StatusCodes.Status409Conflict);
        }

        return TypedResults.Created($"/bank-accounts/{bankAccount.Id}", ToDto(bankAccount));
    }

    private static async Task<Results<Ok<BankAccountDto>, NotFound>> GetById(
        int id, FinanceDbContext db, CancellationToken cancellationToken)
    {
        var bankAccount = await db.BankAccounts.FindAsync([id], cancellationToken);

        return bankAccount is null
            ? TypedResults.NotFound()
            : TypedResults.Ok(ToDto(bankAccount));
    }

    private static bool IsDuplicate(DbUpdateException exception) =>
        exception.InnerException is SqlException { Number: UniqueIndexViolation or UniqueConstraintViolation };

    private static BankAccountDto ToDto(BankAccount bankAccount) => new(
        bankAccount.Id,
        bankAccount.Name,
        bankAccount.BankCode,
        bankAccount.Branch,
        bankAccount.AccountNumber,
        bankAccount.OpeningBalance,
        bankAccount.CreatedAt);
}
