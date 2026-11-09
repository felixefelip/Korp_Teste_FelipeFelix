using Finance.Domain;

namespace Finance.Tests.Domain;

public class CustomerTests
{
    [Theory]
    [InlineData("12345678901")]
    [InlineData("12345678000195")]
    public void AcceptsCpfOrCnpj(string document)
    {
        var customer = new Customer("Acme Ltda", document);

        Assert.Equal("Acme Ltda", customer.Name);
        Assert.Equal(document, customer.Document);
    }

    [Theory]
    [InlineData("")]
    [InlineData("1234567890")]
    [InlineData("123456789012")]
    [InlineData("123.456.789-01")]
    [InlineData("1234567890a")]
    public void RejectsInvalidDocument(string document)
    {
        Assert.Throws<ArgumentException>(() => new Customer("Acme Ltda", document));
    }

    [Theory]
    [InlineData("")]
    [InlineData("   ")]
    public void RejectsBlankName(string name)
    {
        Assert.ThrowsAny<ArgumentException>(() => new Customer(name, "12345678000195"));
    }
}
