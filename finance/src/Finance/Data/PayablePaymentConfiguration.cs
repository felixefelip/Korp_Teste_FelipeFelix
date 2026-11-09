using Finance.Domain;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Finance.Data;

public class PayablePaymentConfiguration : IEntityTypeConfiguration<PayablePayment>
{
    public void Configure(EntityTypeBuilder<PayablePayment> builder)
    {
        builder.ToTable("PayablePayments", table => table.HasCheckConstraint(
            "CK_PayablePayments_Amount", "[Amount] > 0"));
        builder.Property(p => p.Amount).HasPrecision(18, 2);
    }
}
