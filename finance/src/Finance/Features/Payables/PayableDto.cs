using Finance.Domain;

namespace Finance.Features.Payables;

public record PayableDto(
    int Id,
    int BankAccountId,
    string Description,
    int SupplierId,
    decimal Amount,
    decimal PaidAmount,
    decimal RemainingAmount,
    DateOnly DueDate,
    EntryStatus Status,
    DateTimeOffset CreatedAt);
