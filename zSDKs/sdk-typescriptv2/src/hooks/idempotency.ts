import { Awaitable, BeforeRequestContext, BeforeRequestHook } from "./types.js";

import { v4 as uuidv4 } from 'uuid';

export class IdempotencyHook implements BeforeRequestHook {
    beforeRequest(_: BeforeRequestContext, request: Request): Awaitable<Request> {
        request.headers.set("Idempotency-Key", uuidv4());

        return request;
    }
}