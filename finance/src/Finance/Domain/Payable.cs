namespace Finance.Domain;

public class Payable
{
    private readonly List<PayablePayment> payments = [];

    public int Id { get; private set; }
    public int BankAccountId { get; private set; }
    public string Description { get; private set; }
    public int SupplierId { get; private set; }
    public decimal Amount { get; private set; }
    public decimal PaidAmount { get; private set; }
    public decimal RemainingAmount => Amount - PaidAmount;
    public DateOnly DueDate { get; private set; }
    public EntryStatus Status { get; private set; } = EntryStatus.Open;
    public DateTimeOffset CreatedAt { get; private set; } = DateTimeOffset.UtcNow;
    public IReadOnlyCollection<PayablePayment> Payments => payments;

    public Payable(int bankAccountId, string description, int supplierId, decimal amount, DateOnly dueDate)
    {
        ArgumentOutOfRangeException.ThrowIfNegativeOrZero(bankAccountId);
        ArgumentOutOfRangeException.ThrowIfNegativeOrZero(supplierId);
        ArgumentException.ThrowIfNullOrWhiteSpace(description);
        ArgumentOutOfRangeException.ThrowIfNegativeOrZero(amount);

        BankAccountId = bankAccountId;
        Description = description;
        SupplierId = supplierId;
        Amount = amount;
        DueDate = dueDate;
    }

    public PayablePayment RegisterPayment(decimal amount, DateOnly paidOn)
    {
        ArgumentOutOfRangeException.ThrowIfNegativeOrZero(amount);
        ArgumentOutOfRangeException.ThrowIfGreaterThan(amount, RemainingAmount);

        var payment = new PayablePayment(amount, paidOn);
        payments.Add(payment);
        PaidAmount += amount;

        if (RemainingAmount == 0)
        {
            Status = EntryStatus.Paid;
        }

        return payment;
    }
}
