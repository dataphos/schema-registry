package net.syntio.validity;

import io.apicurio.registry.rules.validity.AvroContentValidator;
import io.apicurio.registry.rules.validity.ContentValidator;
import io.apicurio.registry.rules.validity.JsonSchemaContentValidator;
import io.apicurio.registry.rules.validity.ProtobufContentValidator;
import io.apicurio.registry.rules.validity.XsdContentValidator;

public class ValidatorFactory {
    public static ContentValidator createValidator(String schema) {
        return switch (schema) {
            case SchemaTypes.JSON -> new JsonSchemaContentValidator();
            case SchemaTypes.AVRO -> new AvroContentValidator();
            case SchemaTypes.PROTOBUF -> new ProtobufContentValidator();
            case SchemaTypes.XML -> new XsdContentValidator();
            default -> null;
        };
    }

}
