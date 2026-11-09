namespace Finance.Features.Payables;

public record PayablePaymentDto(
    int Id,
    int PayableId,
    decimal Amount,
    DateOnly PaidOn,
    DateTimeOffset CreatedAt);
