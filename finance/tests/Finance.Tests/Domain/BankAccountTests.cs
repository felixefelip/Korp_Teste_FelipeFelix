using Finance.Domain;

namespace Finance.Tests.Domain;

public class BankAccountTests
{
    [Fact]
    public void KeepsTheGivenData()
    {
        var account = new BankAccount("Itau operacional", "341", "1234", "56789-0", 1500.75m);

        Assert.Equal("Itau operacional", account.Name);
        Assert.Equal("341", account.BankCode);
        Assert.Equal("1234", account.Branch);
        Assert.Equal("56789-0", account.AccountNumber);
        Assert.Equal(1500.75m, account.OpeningBalance);
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-250)]
    public void AcceptsZeroOrNegativeOpeningBalance(decimal openingBalance)
    {
        var account = new BankAccount("Itau operacional", "341", "1234", "56789-0", openingBalance);

        Assert.Equal(openingBalance, account.OpeningBalance);
    }

    [Theory]
    [InlineData("")]
    [InlineData("34")]
    [InlineData("3411")]
    [InlineData("34a")]
    public void RejectsBankCodeThatIsNotThreeDigits(string bankCode)
    {
        Assert.Throws<ArgumentException>(
            () => new BankAccount("Itau operacional", bankCode, "1234", "56789-0", 0m));
    }

    [Theory]
    [InlineData("", "1234", "56789-0")]
    [InlineData("Itau operacional", " ", "56789-0")]
    [InlineData("Itau operacional", "1234", "")]
    public void RejectsBlankFields(string name, string branch, string accountNumber)
    {
        Assert.ThrowsAny<ArgumentException>(
            () => new BankAccount(name, "341", branch, accountNumber, 0m));
    }
}
