using Finance.Domain;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Finance.Data;

public class ReceivablePaymentConfiguration : IEntityTypeConfiguration<ReceivablePayment>
{
    public void Configure(EntityTypeBuilder<ReceivablePayment> builder)
    {
        builder.ToTable("ReceivablePayments", table => table.HasCheckConstraint(
            "CK_ReceivablePayments_Amount", "[Amount] > 0"));
        builder.Property(p => p.Amount).HasPrecision(18, 2);
    }
}
