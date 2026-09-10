import { Faker } from "@faker-js/faker";

const exclusions = /cock/;

// Intercept faker methods and check for exclusions
function withExclusions(faker: Faker) {
  return new Proxy(faker, {
    get(target, name) {
      switch (typeof target[name]) {
        case "function":
          // Wrap the leaf functions and check their outputs
          return function () {
            let res = target[name].apply(target, arguments);
            if (typeof res === "string") {
              while (res.match(exclusions) !== null) {
                res = target[name].apply(target, arguments);
              }
            }

            return res;
          };
        // Faker is used like `faker.word.noun()`, so we need to recurse on the child objects
        case "object":
          return withExclusions(target[name]);
      }

      return target[name];
    },
  });
}

// @ts-ignore
faker = withExclusions(faker.faker);
faker.seed(0);

// Required to ensure that any generated dates use a fixed reference date instead of the time of generation
faker.setDefaultRefDate("2025-01-01T00:00:00.000Z");
