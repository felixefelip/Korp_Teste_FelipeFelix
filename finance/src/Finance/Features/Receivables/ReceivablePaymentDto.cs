namespace Finance.Features.Receivables;

public record ReceivablePaymentDto(
    int Id,
    int ReceivableId,
    decimal Amount,
    DateOnly PaidOn,
    DateTimeOffset CreatedAt);
