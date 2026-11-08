using Finance.Domain;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Finance.Data;

public class PayableConfiguration : IEntityTypeConfiguration<Payable>
{
    public void Configure(EntityTypeBuilder<Payable> builder)
    {
        builder.ToTable("Payables", table => table.HasCheckConstraint(
            "CK_Payables_PaidAmount", "[PaidAmount] >= 0 AND [PaidAmount] <= [Amount]"));
        builder.Property(p => p.Description).HasMaxLength(200);
        builder.Property(p => p.SupplierName).HasMaxLength(150);
        builder.Property(p => p.Amount).HasPrecision(18, 2);
        builder.Property(p => p.PaidAmount).HasPrecision(18, 2);
        builder.Property(p => p.Status).HasConversion<string>().HasMaxLength(20);
        builder.HasIndex(p => p.DueDate);
    }
}
