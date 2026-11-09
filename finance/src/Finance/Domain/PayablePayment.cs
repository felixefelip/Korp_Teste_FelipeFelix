namespace Finance.Domain;

public class PayablePayment
{
    public int Id { get; private set; }
    public int PayableId { get; private set; }
    public decimal Amount { get; private set; }
    public DateOnly PaidOn { get; private set; }
    public DateTimeOffset CreatedAt { get; private set; } = DateTimeOffset.UtcNow;

    internal PayablePayment(decimal amount, DateOnly paidOn)
    {
        Amount = amount;
        PaidOn = paidOn;
    }
}
