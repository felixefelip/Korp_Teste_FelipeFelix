using Finance.Domain;
using Microsoft.EntityFrameworkCore;

namespace Finance.Data;

public class FinanceDbContext(DbContextOptions<FinanceDbContext> options) : DbContext(options)
{
    public DbSet<Receivable> Receivables => Set<Receivable>();
    public DbSet<ReceivablePayment> ReceivablePayments => Set<ReceivablePayment>();
    public DbSet<Payable> Payables => Set<Payable>();
    public DbSet<PayablePayment> PayablePayments => Set<PayablePayment>();
    public DbSet<BankAccount> BankAccounts => Set<BankAccount>();
    public DbSet<Customer> Customers => Set<Customer>();
    public DbSet<Supplier> Suppliers => Set<Supplier>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.ApplyConfigurationsFromAssembly(typeof(FinanceDbContext).Assembly);
    }
}
