using Finance.Data;
using Finance.Domain;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.Data.SqlClient;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;

namespace Finance.Tests.Infrastructure;

public class FinanceApiFactory : WebApplicationFactory<Program>, IAsyncLifetime
{
    private const string TestDatabase = "finance_test";

    protected override void ConfigureWebHost(IWebHostBuilder builder)
    {
        builder.UseSetting("ConnectionStrings:Finance", BuildConnectionString());
    }

    public async Task InitializeAsync()
    {
        await using var scope = Services.CreateAsyncScope();
        var db = scope.ServiceProvider.GetRequiredService<FinanceDbContext>();

        var database = db.Database.GetDbConnection().Database;
        if (database != TestDatabase)
        {
            throw new InvalidOperationException($"Tests must run against {TestDatabase}, not {database}.");
        }

        await db.Database.MigrateAsync();
    }

    public async Task ResetDatabaseAsync()
    {
        await using var scope = Services.CreateAsyncScope();
        var db = scope.ServiceProvider.GetRequiredService<FinanceDbContext>();

        await db.Receivables.ExecuteDeleteAsync();
        await db.Payables.ExecuteDeleteAsync();
        await db.BankAccounts.ExecuteDeleteAsync();
        await db.Customers.ExecuteDeleteAsync();
        await db.Suppliers.ExecuteDeleteAsync();
    }

    public async Task<int> CreateBankAccountAsync()
    {
        await using var scope = Services.CreateAsyncScope();
        var db = scope.ServiceProvider.GetRequiredService<FinanceDbContext>();

        var bankAccount = new BankAccount("Itau operacional", "341", "1234", "56789-0", 0m);
        db.BankAccounts.Add(bankAccount);
        await db.SaveChangesAsync();

        return bankAccount.Id;
    }

    public async Task<int> CreateCustomerAsync(string document = "12345678000195")
    {
        await using var scope = Services.CreateAsyncScope();
        var db = scope.ServiceProvider.GetRequiredService<FinanceDbContext>();

        var customer = new Customer("Acme Ltda", document);
        db.Customers.Add(customer);
        await db.SaveChangesAsync();

        return customer.Id;
    }

    public async Task<int> CreateSupplierAsync(string document = "98765432000110")
    {
        await using var scope = Services.CreateAsyncScope();
        var db = scope.ServiceProvider.GetRequiredService<FinanceDbContext>();

        var supplier = new Supplier("Imobiliaria Central", document);
        db.Suppliers.Add(supplier);
        await db.SaveChangesAsync();

        return supplier.Id;
    }

    Task IAsyncLifetime.DisposeAsync() => DisposeAsync().AsTask();

    private static string BuildConnectionString()
    {
        var connectionString = Environment.GetEnvironmentVariable("ConnectionStrings__Finance")
            ?? throw new InvalidOperationException(
                "ConnectionStrings__Finance is not set. Run the tests inside the finance container.");

        return new SqlConnectionStringBuilder(connectionString) { InitialCatalog = TestDatabase }.ConnectionString;
    }
}
