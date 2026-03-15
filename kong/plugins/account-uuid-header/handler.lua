local cjson = require "cjson.safe"

local AccountUuidHeader = {
  VERSION = "1.0.0",
  PRIORITY = 1449,
}

local function base64url_decode(input)
  if not input then
    return nil
  end

  input = input:gsub("-", "+"):gsub("_", "/")

  local padding = #input % 4
  if padding == 2 then
    input = input .. "=="
  elseif padding == 3 then
    input = input .. "="
  elseif padding == 1 then
    return nil
  end

  return ngx.decode_base64(input)
end

local function get_jwt_payload(raw_token)
  if type(raw_token) ~= "string" then
    return nil, "token is not a string"
  end

  local header_b64, payload_b64, signature_b64 =
    raw_token:match("^([^.]+)%.([^.]+)%.([^.]+)$")

  if not header_b64 or not payload_b64 or not signature_b64 then
    return nil, "invalid jwt format"
  end

  local payload_json = base64url_decode(payload_b64)
  if not payload_json then
    return nil, "failed to decode payload"
  end

  local payload, err = cjson.decode(payload_json)
  if not payload then
    return nil, "failed to parse payload json: " .. tostring(err)
  end

  return payload, nil
end

function AccountUuidHeader:access(conf)
  kong.service.request.clear_header("X-Account-Uuid")
  local raw_token = kong.ctx.shared.authenticated_jwt_token
  if not raw_token then
    kong.log.debug("[account-uuid-header] authenticated_jwt_token not found")
    return
  end

  local payload, err = get_jwt_payload(raw_token)
  if not payload then
    kong.log.err("[account-uuid-header] ", tostring(err))
    return
  end

  local account_uuid = payload.sub
  if not account_uuid or account_uuid == "" then
    kong.log.debug("[account-uuid-header] claim 'sub' not found")
    return
  end

  kong.log.err("[account-uuid-header] sub = ", tostring(account_uuid))
  kong.service.request.set_header("X-Account-Uuid", account_uuid)
end

return AccountUuidHeader
