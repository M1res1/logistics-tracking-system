package org.auth.dto;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.Builder;

import java.time.LocalDateTime;
import java.util.Map;

@Builder
@Schema(description = "Standardized error response")
public record ErrorResponse(
    @Schema(description = "Timestamp of the error", example = "2024-05-16T14:33:00")
    LocalDateTime timestamp,
    @Schema(description = "HTTP status code", example = "400")
    int status,
    @Schema(description = "Error category/type", example = "Bad Request")
    String error,
    @Schema(description = "Descriptive error message", example = "Validation failed")
    String message,
    @Schema(description = "Requested API path", example = "/api/v1/auth/register")
    String path,
    @Schema(description = "Map of field-specific validation errors", example = "{\"email\": \"Invalid email format\"}")
    Map<String, String> validationErrors
) {}
