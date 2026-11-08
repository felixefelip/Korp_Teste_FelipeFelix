using Finance.Domain;

namespace Finance.Tests.Domain;

public class ReceivableTests
{
    private static readonly DateOnly DueDate = new(2026, 12, 5);

    [Fact]
    public void StartsOpen()
    {
        var receivable = new Receivable("Venda de 3 notebooks", "Acme Ltda", 4200m, DueDate);

        Assert.Equal(EntryStatus.Open, receivable.Status);
    }

    [Fact]
    public void StartsWithNothingPaid()
    {
        var receivable = new Receivable("Venda de 3 notebooks", "Acme Ltda", 4200.50m, DueDate);

        Assert.Equal(0m, receivable.PaidAmount);
        Assert.Equal(4200.50m, receivable.RemainingAmount);
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-0.01)]
    public void RejectsNonPositiveAmount(decimal amount)
    {
        Assert.Throws<ArgumentOutOfRangeException>(
            () => new Receivable("Venda de 3 notebooks", "Acme Ltda", amount, DueDate));
    }

    [Theory]
    [InlineData("")]
    [InlineData("   ")]
    public void RejectsBlankDescription(string description)
    {
        Assert.ThrowsAny<ArgumentException>(
            () => new Receivable(description, "Acme Ltda", 4200m, DueDate));
    }

    [Theory]
    [InlineData("")]
    [InlineData("   ")]
    public void RejectsBlankCustomerName(string customerName)
    {
        Assert.ThrowsAny<ArgumentException>(
            () => new Receivable("Venda de 3 notebooks", customerName, 4200m, DueDate));
    }
}
