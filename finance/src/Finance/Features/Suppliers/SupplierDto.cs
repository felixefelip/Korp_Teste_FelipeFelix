namespace Finance.Features.Suppliers;

public record SupplierDto(
    int Id,
    string Name,
    string Document,
    DateTimeOffset CreatedAt);
