namespace Finance.Domain;

public class ReceivablePayment
{
    public int Id { get; private set; }
    public int ReceivableId { get; private set; }
    public decimal Amount { get; private set; }
    public DateOnly PaidOn { get; private set; }
    public DateTimeOffset CreatedAt { get; private set; } = DateTimeOffset.UtcNow;

    internal ReceivablePayment(decimal amount, DateOnly paidOn)
    {
        Amount = amount;
        PaidOn = paidOn;
    }
}
