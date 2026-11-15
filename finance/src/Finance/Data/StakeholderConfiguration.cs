using Finance.Domain;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Finance.Data;

public class StakeholderConfiguration : IEntityTypeConfiguration<Stakeholder>
{
    public void Configure(EntityTypeBuilder<Stakeholder> builder)
    {
        builder.ToTable("Stakeholders");
        builder.Property(s => s.Name).HasMaxLength(150);
        builder.Property(s => s.Document).HasMaxLength(14).IsUnicode(false);
        builder.Property(s => s.Type).HasConversion<string>().HasMaxLength(20);
        builder.HasIndex(s => new { s.Type, s.Document }).IsUnique();
    }
}
