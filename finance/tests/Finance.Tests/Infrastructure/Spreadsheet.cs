using System.Net.Http.Headers;
using ClosedXML.Excel;

namespace Finance.Tests.Infrastructure;

public static class Spreadsheet
{
    private const string XlsxContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet";

    public static MultipartFormDataContent Upload(string[] headers, params object[][] rows) =>
        Upload(Build(headers, rows), "registrations.xlsx");

    public static MultipartFormDataContent Upload(byte[] content, string fileName)
    {
        var file = new ByteArrayContent(content);
        file.Headers.ContentType = new MediaTypeHeaderValue(XlsxContentType);

        return new MultipartFormDataContent { { file, "file", fileName } };
    }

    public static byte[] Build(string[] headers, params object[][] rows)
    {
        using var workbook = new XLWorkbook();
        var sheet = workbook.AddWorksheet("Sheet1");

        for (var column = 0; column < headers.Length; column++)
        {
            sheet.Cell(1, column + 1).Value = headers[column];
        }

        for (var row = 0; row < rows.Length; row++)
        {
            for (var column = 0; column < rows[row].Length; column++)
            {
                sheet.Cell(row + 2, column + 1).Value = rows[row][column] switch
                {
                    string text => text,
                    double number => number,
                    var other => throw new ArgumentException($"Unsupported cell value {other}.")
                };
            }
        }

        using var stream = new MemoryStream();
        workbook.SaveAs(stream);
        return stream.ToArray();
    }
}
