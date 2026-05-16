package org.auth.dto;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;
import jakarta.validation.constraints.Size;

@Schema(description = "Registration request payload")
public record RegisterRequest(
    @Email(message = "Invalid email format")
    @NotBlank(message = "Email is required")
    @Schema(description = "User email address", example = "user@example.com")
    String email,
    
    @NotBlank(message = "Password is required")
    @Size(min = 8, message = "Password must be at least 8 characters long")
    @Schema(description = "User password", example = "securePassword123", minLength = 8)
    String password,
    
    @Pattern(regexp = "^\\+?[0-9]{10,15}$", message = "Invalid phone number format")
    @Schema(description = "User phone number", example = "+1234567890")
    String phone,
    
    @NotBlank(message = "User type is required")
    @Pattern(regexp = "^(CUSTOMER|DRIVER|ADMIN|RESTAURANT)$", message = "Invalid user type")
    @Schema(description = "User type/role", allowableValues = {"CUSTOMER", "DRIVER", "ADMIN", "RESTAURANT"})
    String userType
) {}
