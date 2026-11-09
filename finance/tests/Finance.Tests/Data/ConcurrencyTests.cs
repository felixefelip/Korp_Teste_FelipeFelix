using Finance.Data;
using Finance.Domain;
using Finance.Tests.Infrastructure;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;

namespace Finance.Tests.Data;

[Collection(ApiCollection.Name)]
public class ConcurrencyTests(FinanceApiFactory factory) : IAsyncLifetime
{
    private static readonly DateOnly DueDate = new(2026, 12, 10);
    private static readonly DateOnly PaidOn = new(2026, 12, 1);

    private int bankAccountId;

    public async Task InitializeAsync()
    {
        await factory.ResetDatabaseAsync();
        bankAccountId = await factory.CreateBankAccountAsync();
    }

    public Task DisposeAsync() => Task.CompletedTask;

    [Fact]
    public async Task ConcurrentReceivablePaymentIsRejected()
    {
        var id = await SaveAsync(db => db.Receivables, new Receivable(bankAccountId, "Venda", "Acme Ltda", 1000m, DueDate));

        await using var first = Scope();
        await using var second = Scope();
        var firstCopy = await Db(first).Receivables.SingleAsync(r => r.Id == id);
        var secondCopy = await Db(second).Receivables.SingleAsync(r => r.Id == id);

        firstCopy.RegisterPayment(600m, PaidOn);
        await Db(first).SaveChangesAsync();
        secondCopy.RegisterPayment(600m, PaidOn);

        await Assert.ThrowsAsync<DbUpdateConcurrencyException>(() => Db(second).SaveChangesAsync());
    }

    [Fact]
    public async Task ConcurrentPayablePaymentIsRejected()
    {
        var id = await SaveAsync(db => db.Payables, new Payable(bankAccountId, "Aluguel", "Imobiliaria", 1000m, DueDate));

        await using var first = Scope();
        await using var second = Scope();
        var firstCopy = await Db(first).Payables.SingleAsync(p => p.Id == id);
        var secondCopy = await Db(second).Payables.SingleAsync(p => p.Id == id);

        firstCopy.RegisterPayment(600m, PaidOn);
        await Db(first).SaveChangesAsync();
        secondCopy.RegisterPayment(600m, PaidOn);

        await Assert.ThrowsAsync<DbUpdateConcurrencyException>(() => Db(second).SaveChangesAsync());
    }

    private AsyncServiceScope Scope() => factory.Services.CreateAsyncScope();

    private static FinanceDbContext Db(AsyncServiceScope scope) =>
        scope.ServiceProvider.GetRequiredService<FinanceDbContext>();

    private async Task<int> SaveAsync<T>(Func<FinanceDbContext, DbSet<T>> set, T entity) where T : class
    {
        await using var scope = Scope();
        var db = Db(scope);
        set(db).Add(entity);
        await db.SaveChangesAsync();
        return (int)db.Entry(entity).Property("Id").CurrentValue!;
    }
}
