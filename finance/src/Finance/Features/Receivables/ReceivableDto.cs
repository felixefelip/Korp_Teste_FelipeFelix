using Finance.Domain;

namespace Finance.Features.Receivables;

public record ReceivableDto(
    int Id,
    string Description,
    string CustomerName,
    decimal Amount,
    DateOnly DueDate,
    EntryStatus Status,
    DateTimeOffset CreatedAt);
