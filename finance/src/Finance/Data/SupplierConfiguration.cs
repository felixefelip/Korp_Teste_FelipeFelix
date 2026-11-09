using Finance.Domain;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Finance.Data;

public class SupplierConfiguration : IEntityTypeConfiguration<Supplier>
{
    public void Configure(EntityTypeBuilder<Supplier> builder)
    {
        builder.ToTable("Suppliers");
        builder.Property(s => s.Name).HasMaxLength(150);
        builder.Property(s => s.Document).HasMaxLength(14).IsUnicode(false);
        builder.HasIndex(s => s.Document).IsUnique();
    }
}
