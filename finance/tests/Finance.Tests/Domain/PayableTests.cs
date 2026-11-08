using Finance.Domain;

namespace Finance.Tests.Domain;

public class PayableTests
{
    private static readonly DateOnly DueDate = new(2026, 12, 5);

    [Fact]
    public void StartsOpen()
    {
        var payable = new Payable("Aluguel de dezembro", "Imobiliaria Central", 4200m, DueDate);

        Assert.Equal(EntryStatus.Open, payable.Status);
    }

    [Fact]
    public void StartsWithNothingPaid()
    {
        var payable = new Payable("Aluguel de dezembro", "Imobiliaria Central", 4200.50m, DueDate);

        Assert.Equal(0m, payable.PaidAmount);
        Assert.Equal(4200.50m, payable.RemainingAmount);
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-0.01)]
    public void RejectsNonPositiveAmount(decimal amount)
    {
        Assert.Throws<ArgumentOutOfRangeException>(
            () => new Payable("Aluguel de dezembro", "Imobiliaria Central", amount, DueDate));
    }

    [Theory]
    [InlineData("")]
    [InlineData("   ")]
    public void RejectsBlankDescription(string description)
    {
        Assert.ThrowsAny<ArgumentException>(
            () => new Payable(description, "Imobiliaria Central", 4200m, DueDate));
    }

    [Theory]
    [InlineData("")]
    [InlineData("   ")]
    public void RejectsBlankSupplierName(string supplierName)
    {
        Assert.ThrowsAny<ArgumentException>(
            () => new Payable("Aluguel de dezembro", supplierName, 4200m, DueDate));
    }
}
