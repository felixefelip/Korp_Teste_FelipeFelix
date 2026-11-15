using Finance.Domain;

namespace Finance.Tests.Domain;

public class StakeholderTests
{
    [Theory]
    [InlineData("12345678901", StakeholderType.Customer)]
    [InlineData("12345678000195", StakeholderType.Supplier)]
    public void AcceptsCpfOrCnpjOfEitherType(string document, StakeholderType type)
    {
        var stakeholder = new Stakeholder("Acme Ltda", document, type);

        Assert.Equal("Acme Ltda", stakeholder.Name);
        Assert.Equal(document, stakeholder.Document);
        Assert.Equal(type, stakeholder.Type);
    }

    [Theory]
    [InlineData("")]
    [InlineData("1234567890")]
    [InlineData("123456789012")]
    [InlineData("123.456.789-01")]
    [InlineData("1234567890a")]
    public void RejectsInvalidDocument(string document)
    {
        Assert.Throws<ArgumentException>(() => new Stakeholder("Acme Ltda", document, StakeholderType.Customer));
    }

    [Theory]
    [InlineData("")]
    [InlineData("   ")]
    public void RejectsBlankName(string name)
    {
        Assert.ThrowsAny<ArgumentException>(() => new Stakeholder(name, "12345678000195", StakeholderType.Customer));
    }

    [Fact]
    public void RejectsUndefinedType()
    {
        Assert.Throws<ArgumentOutOfRangeException>(
            () => new Stakeholder("Acme Ltda", "12345678000195", (StakeholderType)99));
    }
}
