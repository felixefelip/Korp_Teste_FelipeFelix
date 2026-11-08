namespace Finance.Domain;

public class Payable
{
    public int Id { get; private set; }
    public int BankAccountId { get; private set; }
    public string Description { get; private set; }
    public string SupplierName { get; private set; }
    public decimal Amount { get; private set; }
    public decimal PaidAmount { get; private set; }
    public decimal RemainingAmount => Amount - PaidAmount;
    public DateOnly DueDate { get; private set; }
    public EntryStatus Status { get; private set; } = EntryStatus.Open;
    public DateTimeOffset CreatedAt { get; private set; } = DateTimeOffset.UtcNow;

    public Payable(int bankAccountId, string description, string supplierName, decimal amount, DateOnly dueDate)
    {
        ArgumentOutOfRangeException.ThrowIfNegativeOrZero(bankAccountId);
        ArgumentException.ThrowIfNullOrWhiteSpace(description);
        ArgumentException.ThrowIfNullOrWhiteSpace(supplierName);
        ArgumentOutOfRangeException.ThrowIfNegativeOrZero(amount);

        BankAccountId = bankAccountId;
        Description = description;
        SupplierName = supplierName;
        Amount = amount;
        DueDate = dueDate;
    }
}
