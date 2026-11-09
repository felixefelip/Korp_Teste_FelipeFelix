using Finance.Domain;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Finance.Data;

public class CustomerConfiguration : IEntityTypeConfiguration<Customer>
{
    public void Configure(EntityTypeBuilder<Customer> builder)
    {
        builder.ToTable("Customers");
        builder.Property(c => c.Name).HasMaxLength(150);
        builder.Property(c => c.Document).HasMaxLength(14).IsUnicode(false);
        builder.HasIndex(c => c.Document).IsUnique();
    }
}
