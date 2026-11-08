using Finance.Domain;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Finance.Data;

public class ReceivableConfiguration : IEntityTypeConfiguration<Receivable>
{
    public void Configure(EntityTypeBuilder<Receivable> builder)
    {
        builder.ToTable("Receivables", table => table.HasCheckConstraint(
            "CK_Receivables_PaidAmount", "[PaidAmount] >= 0 AND [PaidAmount] <= [Amount]"));
        builder.Property(r => r.Description).HasMaxLength(200);
        builder.Property(r => r.CustomerName).HasMaxLength(150);
        builder.Property(r => r.Amount).HasPrecision(18, 2);
        builder.Property(r => r.PaidAmount).HasPrecision(18, 2);
        builder.Property(r => r.Status).HasConversion<string>().HasMaxLength(20);
        builder.HasIndex(r => r.DueDate);
        builder.HasOne<BankAccount>()
            .WithMany()
            .HasForeignKey(r => r.BankAccountId)
            .OnDelete(DeleteBehavior.Restrict);
    }
}
