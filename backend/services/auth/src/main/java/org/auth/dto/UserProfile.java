package org.auth.dto;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.Builder;

@Builder
@Schema(description = "User profile information")
public record UserProfile(
    @Schema(description = "Unique identifier of the user", example = "1")
    Long id,
    
    @Schema(description = "User email address", example = "user@example.com")
    String email,
    
    @Schema(description = "User phone number", example = "+1234567890")
    String phone,
    
    @Schema(description = "User type/role", example = "CUSTOMER")
    String userType
) {}
