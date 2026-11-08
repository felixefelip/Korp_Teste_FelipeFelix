namespace Finance.Domain;

public class Receivable
{
    public int Id { get; private set; }
    public int BankAccountId { get; private set; }
    public string Description { get; private set; }
    public string CustomerName { get; private set; }
    public decimal Amount { get; private set; }
    public decimal PaidAmount { get; private set; }
    public decimal RemainingAmount => Amount - PaidAmount;
    public DateOnly DueDate { get; private set; }
    public EntryStatus Status { get; private set; } = EntryStatus.Open;
    public DateTimeOffset CreatedAt { get; private set; } = DateTimeOffset.UtcNow;

    public Receivable(int bankAccountId, string description, string customerName, decimal amount, DateOnly dueDate)
    {
        ArgumentOutOfRangeException.ThrowIfNegativeOrZero(bankAccountId);
        ArgumentException.ThrowIfNullOrWhiteSpace(description);
        ArgumentException.ThrowIfNullOrWhiteSpace(customerName);
        ArgumentOutOfRangeException.ThrowIfNegativeOrZero(amount);

        BankAccountId = bankAccountId;
        Description = description;
        CustomerName = customerName;
        Amount = amount;
        DueDate = dueDate;
    }
}
