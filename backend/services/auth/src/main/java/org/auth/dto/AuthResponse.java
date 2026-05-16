package org.auth.dto;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.Builder;

@Builder
@Schema(description = "Authentication response containing tokens")
public record AuthResponse(
    @Schema(description = "JWT access token")
    String accessToken,
    
    @Schema(description = "JWT refresh token used to obtain new access tokens")
    String refreshToken
) {}
