namespace Finance.Domain;

public class BankAccount
{
    public int Id { get; private set; }
    public string Name { get; private set; }
    public string BankCode { get; private set; }
    public string Branch { get; private set; }
    public string AccountNumber { get; private set; }
    public decimal OpeningBalance { get; private set; }
    public DateTimeOffset CreatedAt { get; private set; } = DateTimeOffset.UtcNow;

    public BankAccount(string name, string bankCode, string branch, string accountNumber, decimal openingBalance)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(name);
        ArgumentException.ThrowIfNullOrWhiteSpace(branch);
        ArgumentException.ThrowIfNullOrWhiteSpace(accountNumber);

        if (bankCode is null || bankCode.Length != 3 || !bankCode.All(char.IsAsciiDigit))
        {
            throw new ArgumentException("Bank code must have exactly 3 digits.", nameof(bankCode));
        }

        Name = name;
        BankCode = bankCode;
        Branch = branch;
        AccountNumber = accountNumber;
        OpeningBalance = openingBalance;
    }
}
