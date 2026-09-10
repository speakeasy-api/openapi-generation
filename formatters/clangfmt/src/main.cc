// Thin stdin/stdout wrapper around clang-format for WASI.
// Reads C# source from stdin, formats it with the Microsoft style, writes to stdout.

#include "clang/Format/Format.h"
#include "clang/Basic/DiagnosticOptions.h"
#include "clang/Rewrite/Core/Rewriter.h"
#include "llvm/Support/raw_ostream.h"

#include <cstdio>
#include <string>
#include <vector>

static std::string readStdin() {
    std::vector<char> buf;
    char chunk[4096];
    while (true) {
        size_t n = fread(chunk, 1, sizeof(chunk), stdin);
        if (n > 0) buf.insert(buf.end(), chunk, chunk + n);
        if (n < sizeof(chunk)) break;
    }
    return std::string(buf.begin(), buf.end());
}

int main(int argc, char **argv) {
    // Determine style: default to Microsoft (closest to idiomatic C#).
    // An optional first argument can override the style name.
    std::string styleName = "microsoft";
    std::string fileName = "input.cs";
    if (argc > 1) styleName = argv[1];
    if (argc > 2) fileName = argv[2];

    std::string code = readStdin();
    if (code.empty()) {
        return 0;
    }

    auto style = clang::format::getStyle(styleName, fileName, "microsoft");
    if (!style) {
        llvm::errs() << "Error: failed to get style '" << styleName << "': "
                      << llvm::toString(style.takeError()) << "\n";
        return 1;
    }

    // Apply the replacements to the source text.
    auto replaces = clang::format::reformat(*style, code,
        clang::tooling::Range(0, code.size()), fileName);

    auto result = clang::tooling::applyAllReplacements(code, replaces);
    if (!result) {
        llvm::errs() << "Error applying replacements: "
                      << llvm::toString(result.takeError()) << "\n";
        return 1;
    }

    llvm::outs() << *result;
    return 0;
}
