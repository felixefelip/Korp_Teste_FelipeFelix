using Finance.Domain;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Finance.Data;

public class BankAccountConfiguration : IEntityTypeConfiguration<BankAccount>
{
    public void Configure(EntityTypeBuilder<BankAccount> builder)
    {
        builder.ToTable("BankAccounts");
        builder.Property(b => b.Name).HasMaxLength(100);
        builder.Property(b => b.BankCode).HasMaxLength(3).IsFixedLength();
        builder.Property(b => b.Branch).HasMaxLength(10);
        builder.Property(b => b.AccountNumber).HasMaxLength(20);
        builder.Property(b => b.OpeningBalance).HasPrecision(18, 2);
        builder.HasIndex(b => new { b.BankCode, b.Branch, b.AccountNumber }).IsUnique();
    }
}
