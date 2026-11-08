using System.ComponentModel.DataAnnotations;

namespace Finance.Features.BankAccounts;

public record CreateBankAccountDto(
    [property: Required, MaxLength(100)] string Name,
    [property: Required, RegularExpression(@"^\d{3}$")] string BankCode,
    [property: Required, MaxLength(10)] string Branch,
    [property: Required, MaxLength(20)] string AccountNumber,
    [property: Required, Range(-999_999_999_999.99, 999_999_999_999.99)] decimal? OpeningBalance);
