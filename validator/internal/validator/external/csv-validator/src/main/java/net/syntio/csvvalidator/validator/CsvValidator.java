package net.syntio.csvvalidator.validator;
import uk.gov.nationalarchives.csv.validator.api.java.FailMessage;
import java.io.Reader;
import java.io.StringReader;
import java.util.ArrayList;
import java.util.List;

public class CsvValidator {
    public static boolean validate(String data, String schema) {
        Reader dataReader = new StringReader(data);
        Reader schemaReader = new StringReader(schema);

        List<FailMessage> messages = uk.gov.nationalarchives.csv.validator.api.java.CsvValidator.validate(dataReader, schemaReader,
                false,
                new ArrayList<>(),
                true,
                false);

        return messages.isEmpty();
    }
}
