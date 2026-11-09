using Finance.Domain;

namespace Finance.Features.Receivables;

public record ReceivableDto(
    int Id,
    int BankAccountId,
    string Description,
    int CustomerId,
    decimal Amount,
    decimal PaidAmount,
    decimal RemainingAmount,
    DateOnly DueDate,
    EntryStatus Status,
    DateTimeOffset CreatedAt);
