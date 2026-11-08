using Finance.Domain;
using Microsoft.EntityFrameworkCore;

namespace Finance.Data;

public class FinanceDbContext(DbContextOptions<FinanceDbContext> options) : DbContext(options)
{
    public DbSet<Receivable> Receivables => Set<Receivable>();
    public DbSet<Payable> Payables => Set<Payable>();
    public DbSet<BankAccount> BankAccounts => Set<BankAccount>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.ApplyConfigurationsFromAssembly(typeof(FinanceDbContext).Assembly);
    }
}
