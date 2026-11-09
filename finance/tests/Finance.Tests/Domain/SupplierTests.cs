using Finance.Domain;

namespace Finance.Tests.Domain;

public class SupplierTests
{
    [Theory]
    [InlineData("12345678901")]
    [InlineData("12345678000195")]
    public void AcceptsCpfOrCnpj(string document)
    {
        var supplier = new Supplier("Imobiliaria Central", document);

        Assert.Equal("Imobiliaria Central", supplier.Name);
        Assert.Equal(document, supplier.Document);
    }

    [Theory]
    [InlineData("")]
    [InlineData("1234567890")]
    [InlineData("123456789012")]
    [InlineData("123.456.789-01")]
    [InlineData("1234567890a")]
    public void RejectsInvalidDocument(string document)
    {
        Assert.Throws<ArgumentException>(() => new Supplier("Imobiliaria Central", document));
    }

    [Theory]
    [InlineData("")]
    [InlineData("   ")]
    public void RejectsBlankName(string name)
    {
        Assert.ThrowsAny<ArgumentException>(() => new Supplier(name, "12345678000195"));
    }
}
