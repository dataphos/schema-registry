package net.syntio.compatibility;

import net.syntio.compatibility.checker.AvroChecker;
import net.syntio.compatibility.checker.CompatibilityChecker;
import net.syntio.compatibility.checker.JsonChecker;
import net.syntio.compatibility.checker.ProtobufChecker;

public class CheckerFactory {
    public static CompatibilityChecker createChecker(String format) throws Exception {
        return switch (format) {
            case FileTypes.JSON -> new JsonChecker();
            case FileTypes.PROTOBUF -> new ProtobufChecker();
            case FileTypes.AVRO -> new AvroChecker();
            default -> throw new Exception("Unknown format");
        };
    }
}
