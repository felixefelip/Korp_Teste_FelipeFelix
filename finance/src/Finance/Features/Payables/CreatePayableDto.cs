using System.ComponentModel.DataAnnotations;

namespace Finance.Features.Payables;

public record CreatePayableDto(
    [property: Required, Range(1, int.MaxValue)] int? BankAccountId,
    [property: Required, MaxLength(200)] string Description,
    [property: Required, Range(1, int.MaxValue)] int? SupplierId,
    [property: Required, Range(0.01, 999_999_999_999.99)] decimal? Amount,
    [property: Required] DateOnly? DueDate);
