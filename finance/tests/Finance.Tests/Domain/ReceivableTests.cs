using Finance.Domain;

namespace Finance.Tests.Domain;

public class ReceivableTests
{
    private const int BankAccountId = 1;
    private static readonly DateOnly DueDate = new(2026, 12, 5);

    [Fact]
    public void StartsOpen()
    {
        var receivable = new Receivable(BankAccountId, "Venda de 3 notebooks", "Acme Ltda", 4200m, DueDate);

        Assert.Equal(EntryStatus.Open, receivable.Status);
    }

    [Fact]
    public void StartsWithNothingPaid()
    {
        var receivable = new Receivable(BankAccountId, "Venda de 3 notebooks", "Acme Ltda", 4200.50m, DueDate);

        Assert.Equal(0m, receivable.PaidAmount);
        Assert.Equal(4200.50m, receivable.RemainingAmount);
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-0.01)]
    public void RejectsNonPositiveAmount(decimal amount)
    {
        Assert.Throws<ArgumentOutOfRangeException>(
            () => new Receivable(BankAccountId, "Venda de 3 notebooks", "Acme Ltda", amount, DueDate));
    }

    [Theory]
    [InlineData("")]
    [InlineData("   ")]
    public void RejectsBlankDescription(string description)
    {
        Assert.ThrowsAny<ArgumentException>(
            () => new Receivable(BankAccountId, description, "Acme Ltda", 4200m, DueDate));
    }

    [Theory]
    [InlineData("")]
    [InlineData("   ")]
    public void RejectsBlankCustomerName(string customerName)
    {
        Assert.ThrowsAny<ArgumentException>(
            () => new Receivable(BankAccountId, "Venda de 3 notebooks", customerName, 4200m, DueDate));
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-1)]
    public void RejectsInvalidBankAccountId(int bankAccountId)
    {
        Assert.Throws<ArgumentOutOfRangeException>(
            () => new Receivable(bankAccountId, "Venda de 3 notebooks", "Acme Ltda", 4200m, DueDate));
    }

    [Fact]
    public void PartialPaymentKeepsItOpen()
    {
        var receivable = NewReceivable(1000m);

        var payment = receivable.RegisterPayment(400m, new DateOnly(2026, 12, 1));

        Assert.Equal(400m, payment.Amount);
        Assert.Equal(new DateOnly(2026, 12, 1), payment.PaidOn);
        Assert.Equal(400m, receivable.PaidAmount);
        Assert.Equal(600m, receivable.RemainingAmount);
        Assert.Equal(EntryStatus.Open, receivable.Status);
        Assert.Single(receivable.Payments);
    }

    [Fact]
    public void PayingTheRemainingMarksItPaid()
    {
        var receivable = NewReceivable(1000m);

        receivable.RegisterPayment(400m, new DateOnly(2026, 12, 1));
        receivable.RegisterPayment(600m, new DateOnly(2026, 12, 2));

        Assert.Equal(1000m, receivable.PaidAmount);
        Assert.Equal(0m, receivable.RemainingAmount);
        Assert.Equal(EntryStatus.Paid, receivable.Status);
        Assert.Equal(2, receivable.Payments.Count);
    }

    [Fact]
    public void RejectsPaymentAboveRemaining()
    {
        var receivable = NewReceivable(1000m);
        receivable.RegisterPayment(400m, new DateOnly(2026, 12, 1));

        Assert.Throws<ArgumentOutOfRangeException>(
            () => receivable.RegisterPayment(600.01m, new DateOnly(2026, 12, 2)));

        Assert.Equal(400m, receivable.PaidAmount);
        Assert.Single(receivable.Payments);
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-1)]
    public void RejectsNonPositivePayment(decimal amount)
    {
        var receivable = NewReceivable(1000m);

        Assert.Throws<ArgumentOutOfRangeException>(
            () => receivable.RegisterPayment(amount, new DateOnly(2026, 12, 1)));
    }

    private static Receivable NewReceivable(decimal amount) =>
        new(BankAccountId, "Venda de 3 notebooks", "Acme Ltda", amount, DueDate);
}
