namespace Finance.Tests.Infrastructure;

[CollectionDefinition(Name)]
public class ApiCollection : ICollectionFixture<FinanceApiFactory>
{
    public const string Name = "Api";
}
