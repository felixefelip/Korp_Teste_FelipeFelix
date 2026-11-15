using System.Globalization;
using ClosedXML.Excel;
using Finance.Domain;

namespace Finance.Features.Spreadsheets;

public record RegistrationRow(int Number, string Name, string Document);

public record RegistrationSpreadsheetResult(
    IReadOnlyList<RegistrationRow> Rows,
    Dictionary<string, string[]> Errors);

public static class RegistrationSpreadsheet
{
    public const long MaxFileSize = 5 * 1024 * 1024;
    public const int MaxRows = 5000;
    public const string NameColumn = "Name";
    public const string DocumentColumn = "Document";
    public const int NameMaxLength = 150;

    private const string FileKey = "File";

    public static RegistrationSpreadsheetResult Read(IFormFile file)
    {
        if (!file.FileName.EndsWith(".xlsx", StringComparison.OrdinalIgnoreCase))
        {
            return Fail("The file must be an .xlsx spreadsheet.");
        }

        if (file.Length == 0)
        {
            return Fail("The file is empty.");
        }

        if (file.Length > MaxFileSize)
        {
            return Fail($"The file must be at most {MaxFileSize / 1024 / 1024} MB.");
        }

        XLWorkbook workbook;

        try
        {
            using var stream = file.OpenReadStream();
            workbook = new XLWorkbook(stream);
        }
        catch (Exception)
        {
            return Fail("The file is not a valid .xlsx spreadsheet.");
        }

        using (workbook)
        {
            return ReadRows(workbook.Worksheet(1));
        }
    }

    public static Dictionary<string, string[]> Validate(
        IReadOnlyList<RegistrationRow> rows, IReadOnlySet<string> registeredDocuments)
    {
        var errors = new Dictionary<string, string[]>();
        var firstRowByDocument = new Dictionary<string, int>();

        foreach (var row in rows)
        {
            var rowErrors = new List<string>();

            if (string.IsNullOrWhiteSpace(row.Name))
            {
                rowErrors.Add($"{NameColumn} is required.");
            }
            else if (row.Name.Length > NameMaxLength)
            {
                rowErrors.Add($"{NameColumn} must be at most {NameMaxLength} characters.");
            }

            if (!TaxDocument.IsValidFormat(row.Document))
            {
                rowErrors.Add($"{DocumentColumn} must be a CPF (11 digits) or a CNPJ (14 digits).");
            }
            else if (registeredDocuments.Contains(row.Document))
            {
                rowErrors.Add($"{DocumentColumn} is already registered.");
            }
            else if (!firstRowByDocument.TryAdd(row.Document, row.Number))
            {
                rowErrors.Add($"{DocumentColumn} is repeated from row {firstRowByDocument[row.Document]}.");
            }

            if (rowErrors.Count > 0)
            {
                errors[$"Row {row.Number}"] = [.. rowErrors];
            }
        }

        return errors;
    }

    private static RegistrationSpreadsheetResult ReadRows(IXLWorksheet sheet)
    {
        var header = sheet.FirstRowUsed();

        if (header is null)
        {
            return Fail("The spreadsheet is empty.");
        }

        var columns = header.CellsUsed()
            .GroupBy(cell => cell.GetString().Trim(), StringComparer.OrdinalIgnoreCase)
            .ToDictionary(group => group.Key, group => group.First().Address.ColumnNumber, StringComparer.OrdinalIgnoreCase);

        var missing = new[] { NameColumn, DocumentColumn }.Where(column => !columns.ContainsKey(column)).ToArray();

        if (missing.Length > 0)
        {
            return Fail($"Missing column(s): {string.Join(", ", missing)}.");
        }

        var dataRows = sheet.RowsUsed().Where(row => row.RowNumber() > header.RowNumber()).ToList();

        if (dataRows.Count == 0)
        {
            return Fail("The spreadsheet has no rows to import.");
        }

        if (dataRows.Count > MaxRows)
        {
            return Fail($"The spreadsheet must have at most {MaxRows} rows.");
        }

        var rows = dataRows
            .Select(row => new RegistrationRow(
                row.RowNumber(),
                row.Cell(columns[NameColumn]).GetString().Trim(),
                NormalizeDocument(row.Cell(columns[DocumentColumn]))))
            .ToList();

        return new RegistrationSpreadsheetResult(rows, []);
    }

    private static string NormalizeDocument(IXLCell cell)
    {
        if (cell.DataType == XLDataType.Number)
        {
            var digits = Math.Truncate(cell.GetDouble()).ToString("0", CultureInfo.InvariantCulture);
            return digits.Length <= 11 ? digits.PadLeft(11, '0') : digits.PadLeft(14, '0');
        }

        return new string(cell.GetString().Where(c => c is not ('.' or '-' or '/' or ' ')).ToArray());
    }

    private static RegistrationSpreadsheetResult Fail(string message) =>
        new([], new Dictionary<string, string[]> { [FileKey] = [message] });
}
