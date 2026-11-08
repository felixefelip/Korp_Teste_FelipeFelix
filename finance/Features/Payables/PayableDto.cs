using Finance.Domain;

namespace Finance.Features.Payables;

public record PayableDto(
    int Id,
    string Description,
    string SupplierName,
    decimal Amount,
    DateOnly DueDate,
    EntryStatus Status,
    DateTimeOffset CreatedAt);
