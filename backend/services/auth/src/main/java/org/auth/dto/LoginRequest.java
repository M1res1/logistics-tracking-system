package org.auth.dto;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;

@Schema(description = "Login request payload")
public record LoginRequest(
    @Email(message = "Invalid email format")
    @NotBlank(message = "Email is required")
    @Schema(description = "User email address", example = "user@example.com")
    String email,
    
    @NotBlank(message = "Password is required")
    @Schema(description = "User password", example = "securePassword123")
    String password,

    @Schema(description = "Optional device information for the session", example = "iPhone 15, iOS 17.0")
    String deviceInfo
) {}
