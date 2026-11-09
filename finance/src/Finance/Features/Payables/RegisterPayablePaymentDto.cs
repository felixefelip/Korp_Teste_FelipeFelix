using System.ComponentModel.DataAnnotations;

namespace Finance.Features.Payables;

public record RegisterPayablePaymentDto(
    [property: Required, Range(0.01, 999_999_999_999.99)] decimal? Amount,
    [property: Required] DateOnly? PaidOn);
