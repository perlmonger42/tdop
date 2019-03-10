// tokens.js
// 2011-01-04

// (c) 2006 Douglas Crockford

// Produce an array of simple token objects from a string.
// A simple token object contains these members:
//      type: 'name', 'string', 'number', 'punctuator'
//      value: string or number value of the token
//      from: index of first character of the token
//      to: index of the last character + 1

// Comments of the // type are ignored.

const rx_crlf = /\n|\r\n?/;

// const rx_crlf = /
//      \n
// |
//      \r \n?
// /;


function tokenize(source) {

// tokenize takes a source and produces from it an array of token objects.
// If the source is not an array, then it is split into lines at the
// carriage return/linefeed.

    const lines = (Array.isArray(source))
        ? source
        : source.split(rx_crlf);
    const result = [];

    lines.forEach(function (line, line_nr) {
        const rx_token = /(\u0020+)|(\/\/.*)|([a-zA-Z][a-zA-Z_0-9]*)|(\d+(?:\.\d+)?(?:[eE][+\-]?\d+)?)|("(?:[^"\\]|\\(?:[nr"\\]|u[0-9a-fA-F]{4}))*")|([(){}\[\]?.,:;~*\/]|&&?|\|\|?|[+\-<>]=?|[!=](?:==)?)/y;

// const rx_token = /
//     ( \u0020+ )
// |
//     ( \/\/ .* )
// |
//     (
//         [ a-z A-Z ]
//         [ a-z A-Z _ 0-9 ]*
//     )
// |
//     (
//         \d+
//         (?: \. \d+ )?
//         (?: [ e E ] [ + \- ]? \d+ )?
//     )
// |
//     (
//         "
//         (?:
//             [^ " \\ ]
//         |
//             \\
//             (?:
//                 [ n r " \\ ]
//             |
//                 u [ 0-9 a-f A-F ]{4}
//             )
//         )*
//         "
//     )
// |
//     (
//         [ ( ) { } \[ \] ? . , : ; ~ * \/ ]
//     |
//         & &?
//     |
//         \| \|?
//     |
//         [ + \-  < > ] =?
//     |
//         [ ! = ] (?: == )?
//     )
// /y;

// Capture Group
// [1]  Whitespace
// [2]  Comment
// [3]  Name
// [4]  Number
// [5]  String
// [6]  Punctuator

        let column_nr = 0;
        let make = function (type, value) {

// Make a token object and append it to the result.

            result.push({
                type,
                value,
                line_nr,
                column_nr
            });
        };

        while (column_nr < line.length) {
            let captives = rx_token.exec(line);
            if (!captives) {
                throw new SyntaxError(
                    "line "
                    + line_nr
                    + " column "
                    + column_nr
                );
            } else if (captives[1]) {
            } else if (captives[2]) {
            } else if (captives[3]) {
                make("name", captives[3]);
            } else if (captives[4]) {
                let number = Number(captives[4]);
                if (Number.isFinite(number)) {
                    make("number", );
                } else {
                    throw new TypeError(
                        "line "
                        + line_nr
                        + " column "
                        + column_nr
                    );
                }
            } else if (captives[5]) {
                make("string", JSON.parse(captives[5]));
            } else if (captives[6]) {
                make("punctuator", captives[6]);
            }
            column_nr = rx_token.lastIndex;
        }
    });
    return result;
}
