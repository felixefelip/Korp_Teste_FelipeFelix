namespace Finance.Domain;

public class Receivable
{
    private readonly List<ReceivablePayment> payments = [];

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
    public IReadOnlyCollection<ReceivablePayment> Payments => payments;

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

    public ReceivablePayment RegisterPayment(decimal amount, DateOnly paidOn)
    {
        ArgumentOutOfRangeException.ThrowIfNegativeOrZero(amount);
        ArgumentOutOfRangeException.ThrowIfGreaterThan(amount, RemainingAmount);

        var payment = new ReceivablePayment(amount, paidOn);
        payments.Add(payment);
        PaidAmount += amount;

        if (RemainingAmount == 0)
        {
            Status = EntryStatus.Paid;
        }

        return payment;
    }
}
