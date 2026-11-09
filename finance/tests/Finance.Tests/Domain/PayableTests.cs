using Finance.Domain;

namespace Finance.Tests.Domain;

public class PayableTests
{
    private const int BankAccountId = 1;
    private const int SupplierId = 2;
    private static readonly DateOnly DueDate = new(2026, 12, 5);

    [Fact]
    public void StartsOpen()
    {
        var payable = new Payable(BankAccountId, "Aluguel de dezembro", SupplierId, 4200m, DueDate);

        Assert.Equal(EntryStatus.Open, payable.Status);
    }

    [Fact]
    public void StartsWithNothingPaid()
    {
        var payable = new Payable(BankAccountId, "Aluguel de dezembro", SupplierId, 4200.50m, DueDate);

        Assert.Equal(0m, payable.PaidAmount);
        Assert.Equal(4200.50m, payable.RemainingAmount);
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-0.01)]
    public void RejectsNonPositiveAmount(decimal amount)
    {
        Assert.Throws<ArgumentOutOfRangeException>(
            () => new Payable(BankAccountId, "Aluguel de dezembro", SupplierId, amount, DueDate));
    }

    [Theory]
    [InlineData("")]
    [InlineData("   ")]
    public void RejectsBlankDescription(string description)
    {
        Assert.ThrowsAny<ArgumentException>(
            () => new Payable(BankAccountId, description, SupplierId, 4200m, DueDate));
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-1)]
    public void RejectsInvalidSupplierId(int supplierId)
    {
        Assert.Throws<ArgumentOutOfRangeException>(
            () => new Payable(BankAccountId, "Aluguel de dezembro", supplierId, 4200m, DueDate));
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-1)]
    public void RejectsInvalidBankAccountId(int bankAccountId)
    {
        Assert.Throws<ArgumentOutOfRangeException>(
            () => new Payable(bankAccountId, "Aluguel de dezembro", SupplierId, 4200m, DueDate));
    }

    [Fact]
    public void PartialPaymentKeepsItOpen()
    {
        var payable = NewPayable(1000m);

        var payment = payable.RegisterPayment(400m, new DateOnly(2026, 12, 1));

        Assert.Equal(400m, payment.Amount);
        Assert.Equal(new DateOnly(2026, 12, 1), payment.PaidOn);
        Assert.Equal(400m, payable.PaidAmount);
        Assert.Equal(600m, payable.RemainingAmount);
        Assert.Equal(EntryStatus.Open, payable.Status);
        Assert.Single(payable.Payments);
    }

    [Fact]
    public void PayingTheRemainingMarksItPaid()
    {
        var payable = NewPayable(1000m);

        payable.RegisterPayment(400m, new DateOnly(2026, 12, 1));
        payable.RegisterPayment(600m, new DateOnly(2026, 12, 2));

        Assert.Equal(1000m, payable.PaidAmount);
        Assert.Equal(0m, payable.RemainingAmount);
        Assert.Equal(EntryStatus.Paid, payable.Status);
        Assert.Equal(2, payable.Payments.Count);
    }

    [Fact]
    public void RejectsPaymentAboveRemaining()
    {
        var payable = NewPayable(1000m);
        payable.RegisterPayment(400m, new DateOnly(2026, 12, 1));

        Assert.Throws<ArgumentOutOfRangeException>(
            () => payable.RegisterPayment(600.01m, new DateOnly(2026, 12, 2)));

        Assert.Equal(400m, payable.PaidAmount);
        Assert.Single(payable.Payments);
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-1)]
    public void RejectsNonPositivePayment(decimal amount)
    {
        var payable = NewPayable(1000m);

        Assert.Throws<ArgumentOutOfRangeException>(
            () => payable.RegisterPayment(amount, new DateOnly(2026, 12, 1)));
    }

    private static Payable NewPayable(decimal amount) =>
        new(BankAccountId, "Aluguel de dezembro", SupplierId, amount, DueDate);
}
