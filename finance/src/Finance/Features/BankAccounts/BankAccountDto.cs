namespace Finance.Features.BankAccounts;

public record BankAccountDto(
    int Id,
    string Name,
    string BankCode,
    string Branch,
    string AccountNumber,
    decimal OpeningBalance,
    DateTimeOffset CreatedAt);
