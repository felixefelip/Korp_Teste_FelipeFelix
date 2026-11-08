using Finance.Domain;

namespace Finance.Features.Receivables;

public record ReceivableDto(
    int Id,
    string Description,
    string CustomerName,
    decimal Amount,
    decimal PaidAmount,
    decimal RemainingAmount,
    DateOnly DueDate,
    EntryStatus Status,
    DateTimeOffset CreatedAt);
