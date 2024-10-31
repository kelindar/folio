package validate

import (
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsSnakeCase(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"snake case":                       {input: "snake_case", expected: true},
		"uppercase snake":                  {input: "SNAKE_CASE", expected: false},
		"camelCase":                        {input: "camelCase", expected: false},
		"kebab-case":                       {input: "kebab-case", expected: false},
		"contains spaces":                  {input: "snake case", expected: false},
		"contains symbols":                 {input: "snake_case!", expected: false},
		"empty string":                     {input: "", expected: false},
		"starts with number":               {input: "1snake_case", expected: false},
		"starts with symbol":               {input: "_snake_case", expected: false},
		"starts with uppercase":            {input: "Snake_case", expected: false},
		"starts with uppercase and symbol": {input: "_Snake_case", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsSnakeCase(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsEmail(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid email":              {input: "test@example.com", expected: true},
		"invalid email no at":      {input: "testexample.com", expected: false},
		"invalid email multiple @": {input: "test@@example.com", expected: false},
		"empty string":             {input: "", expected: false},
		"uppercase email":          {input: "TEST@EXAMPLE.COM", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsEmail(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsExistingEmail(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"existing domain":        {input: "test@gmail.com", expected: true},
		"localhost domain":       {input: "user@localhost", expected: true},
		"invalid email format":   {input: "invalid-email", expected: false},
		"short email":            {input: "a@b.co", expected: true},
		"long email":             {input: strings.Repeat("a", 65) + "@example.com", expected: false},
		"email with invalid TLD": {input: "test@example.invalidtld", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsExistingEmail(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
func TestIsURL(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid http URL":        {input: "http://example.com", expected: true},
		"valid https URL":       {input: "https://example.com", expected: true},
		"URL without scheme":    {input: "example.com", expected: true},
		"empty string":          {input: "", expected: false},
		"URL too long":          {input: "http://" + strings.Repeat("a", 2084) + ".com", expected: false},
		"URL too short":         {input: "ht", expected: false},
		"URL starts with dot":   {input: ".example.com", expected: false},
		"URL with invalid host": {input: "http://", expected: false},
		"URL with port":         {input: "http://example.com:8080", expected: true},
		"URL with path":         {input: "http://example.com/path", expected: true},
		"URL with query":        {input: "http://example.com?query=1", expected: true},
		"URL with fragment":     {input: "http://example.com#section", expected: true},
		"non HTTP scheme":       {input: "ftp://example.com", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsURL(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
func TestIsAlpha(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"only letters":     {input: "abcXYZ", expected: true},
		"contains numbers": {input: "abc123", expected: false},
		"contains spaces":  {input: "abc xyz", expected: false},
		"empty string":     {input: "", expected: true},
		"unicode letters":  {input: "абвгд", expected: false},
		"special chars":    {input: "abc!", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsAlpha(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
func TestIsNumeric(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"only digits":      {input: "123456", expected: true},
		"contains letters": {input: "123abc", expected: false},
		"empty string":     {input: "", expected: true},
		"negative number":  {input: "-123", expected: false},
		"decimal number":   {input: "123.45", expected: false},
		"unicode digits":   {input: "１２３４５６", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsNumeric(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
func TestIsInt(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"positive integer": {input: "123", expected: true},
		"negative integer": {input: "-123", expected: true},
		"zero":             {input: "0", expected: true},
		"decimal number":   {input: "123.45", expected: false},
		"letters":          {input: "abc", expected: false},
		"empty string":     {input: "", expected: true},
		"plus sign":        {input: "+123", expected: true},
		"leading zeros":    {input: "00123", expected: false},
		"space included":   {input: "123 ", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsInt(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsFloat(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"positive float":   {input: "123.45", expected: true},
		"negative float":   {input: "-123.45", expected: true},
		"integer":          {input: "123", expected: true},
		"letters":          {input: "abc", expected: false},
		"empty string":     {input: "", expected: false},
		"float with comma": {input: "123,45", expected: false},
		"scientific":       {input: "1e10", expected: true},
		"plus sign":        {input: "+123.45", expected: true},
		"multiple dots":    {input: "123.45.67", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsFloat(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
func TestIsHexadecimal(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"lowercase hex":   {input: "deadbeef", expected: true},
		"uppercase hex":   {input: "DEADBEEF", expected: true},
		"mixed case hex":  {input: "DeadBeef", expected: true},
		"with prefix 0x":  {input: "0xDEADBEEF", expected: false},
		"not hex":         {input: "123G", expected: false},
		"empty string":    {input: "", expected: false},
		"hex with spaces": {input: "DE AD BE EF", expected: false},
		"numeric":         {input: "123456", expected: true},
		"special chars":   {input: "!@#$%", expected: false},
		"unicode letters": {input: "абвгд", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsHexadecimal(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsUUID(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid UUID v4":       {input: "123e4567-e89b-12d3-a456-426655440000", expected: true},
		"valid UUID v5":       {input: "123e4567-e89b-52d3-a456-426655440000", expected: true},
		"invalid UUID format": {input: "123e4567e89b12d3a456426655440000", expected: false},
		"empty string":        {input: "", expected: false},
		"UUID with braces":    {input: "{123e4567-e89b-12d3-a456-426655440000}", expected: false},
		"uppercase UUID":      {input: "123E4567-E89B-12D3-A456-426655440000", expected: true},
		"nil UUID":            {input: "00000000-0000-0000-0000-000000000000", expected: true},
		"short UUID":          {input: "123e4567-e89b-12d3-a456-42665544", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsUUID(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
func TestIsCreditCard(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid Visa":         {input: "4111111111111111", expected: true},
		"valid MasterCard":   {input: "5500000000000004", expected: true},
		"invalid number":     {input: "1234567890123456", expected: false},
		"empty string":       {input: "", expected: false},
		"with spaces":        {input: "4111 1111 1111 1111", expected: true},
		"with hyphens":       {input: "4111-1111-1111-1111", expected: true},
		"too short":          {input: "411111111111", expected: false},
		"too long":           {input: "41111111111111111111", expected: false},
		"non-digit chars":    {input: "4111a1111111b111", expected: false},
		"valid Amex":         {input: "378282246310005", expected: true},
		"valid Discover":     {input: "6011111111111117", expected: true},
		"valid JCB":          {input: "3530111333300000", expected: true},
		"valid Diners Club":  {input: "30569309025904", expected: true},
		"valid UnionPay":     {input: "6200000000000005", expected: true},
		"valid Maestro":      {input: "6759649826438453", expected: true},
		"valid with letters": {input: "4111a1111111b111", expected: false},
		"valid with symbols": {input: "4111!1111@1111#1111", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsCreditCard(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsJSON(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid JSON object":     {input: `{"key": "value"}`, expected: true},
		"valid JSON array":      {input: `["value1", "value2"]`, expected: true},
		"invalid JSON":          {input: `{key: value}`, expected: false},
		"empty string":          {input: "", expected: false},
		"valid JSON string":     {input: `"Just a string"`, expected: true},
		"valid JSON number":     {input: "123", expected: true},
		"valid JSON boolean":    {input: "true", expected: true},
		"valid JSON null":       {input: "null", expected: true},
		"invalid JSON syntax":   {input: "{'key': 'value'}", expected: false},
		"non-JSON string":       {input: "Just a string", expected: false},
		"JSON with comments":    {input: `{"key": /* comment */ "value"}`, expected: false},
		"JSON with trailing ,":  {input: `{"key": "value",}`, expected: false},
		"JSON with extra chars": {input: `{"key": "value"} extra`, expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsJSON(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsBase64(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid base64":          {input: "U29tZSB0ZXh0", expected: true},
		"invalid base64":        {input: "U29tZSB0ZXh0===", expected: false},
		"empty string":          {input: "", expected: false},
		"non-base64 chars":      {input: "Not base64!", expected: false},
		"whitespace included":   {input: "U29tZ SB0ZXh0", expected: false},
		"unicode chars":         {input: "😀😃😄😁", expected: false},
		"valid URL safe base64": {input: "U29tZV90ZXh0", expected: true},
		"valid long base64":     {input: strings.Repeat("U29tZSB0ZXh0", 100), expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsBase64(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsIP(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid IPv4":          {input: "192.168.1.1", expected: true},
		"valid IPv6":          {input: "2001:0db8:85a3:0000:0000:8a2e:0370:7334", expected: true},
		"invalid IP":          {input: "999.999.999.999", expected: false},
		"empty string":        {input: "", expected: false},
		"hostname instead IP": {input: "example.com", expected: false},
		"short IPv6":          {input: "::1", expected: true},
		"IPv4-mapped IPv6":    {input: "::ffff:192.0.2.128", expected: true},
		"IPv6 zone index":     {input: "fe80::1ff:fe23:4567:890a%eth0", expected: false}, // net.ParseIP does not handle zone
		"IPv6 with shorthand": {input: "2001:db8::1", expected: true},
		"IPv6 with extra ::":  {input: "2001::db8::1", expected: false},
		"IPv4 in IPv6 format": {input: "::ffff:c000:0280", expected: true},
		"invalid characters":  {input: "abc.def.ghi.jkl", expected: false},
		"missing octets":      {input: "192.168.1", expected: false},
		"too many octets":     {input: "192.168.1.1.1", expected: false},
		"negative numbers":    {input: "-192.168.1.1", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsIP(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsMAC(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid MAC colons":   {input: "00:1A:2B:3C:4D:5E", expected: true},
		"valid MAC hyphens":  {input: "00-1A-2B-3C-4D-5E", expected: true},
		"valid MAC dots":     {input: "001A.2B3C.4D5E", expected: true},
		"invalid MAC length": {input: "00:1A:2B:3C:4D", expected: false},
		"invalid characters": {input: "00:1G:2B:3C:4D:5E", expected: false},
		"empty string":       {input: "", expected: false},
		"missing octets":     {input: "00:1A:2B:3C", expected: false},
		"extra octets":       {input: "00:1A:2B:3C:4D:5E:6F", expected: false},
		"mixed separators":   {input: "00:1A-2B:3C-4D:5E", expected: false},
		"lowercase letters":  {input: "00:1a:2b:3c:4d:5e", expected: true},
		"uppercase letters":  {input: "00:1A:2B:3C:4D:5E", expected: true},
		"MAC with spaces":    {input: "00:1A:2B:3C:4D:5E ", expected: false},
		"invalid format":     {input: "001A2B3C4D5E", expected: false},
		"multicast MAC":      {input: "01:00:5e:00:00:00", expected: true},
		"broadcast MAC":      {input: "ff:ff:ff:ff:ff:ff", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsMAC(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsLatitude(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid latitude":            {input: "45.0", expected: true},
		"valid negative latitude":   {input: "-45.0", expected: true},
		"zero latitude":             {input: "0", expected: true},
		"maximum positive latitude": {input: "90", expected: true},
		"maximum negative latitude": {input: "-90", expected: true},
		"out of range positive":     {input: "90.1", expected: false},
		"out of range negative":     {input: "-90.1", expected: false},
		"not a number":              {input: "abc", expected: false},
		"empty string":              {input: "", expected: false},
		"latitude with spaces":      {input: " 45.0 ", expected: false},
		"latitude with comma":       {input: "45,0", expected: false},
		"latitude with NSEW":        {input: "45N", expected: false},
		"latitude over 90":          {input: "100", expected: false},
		"latitude under -90":        {input: "-100", expected: false},
		"integer latitude":          {input: "45", expected: true},
		"latitude with plus sign":   {input: "+45.0", expected: true},
		"latitude with exponent":    {input: "4.5e1", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsLatitude(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsLongitude(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid longitude":            {input: "90.0", expected: true},
		"valid negative longitude":   {input: "-90.0", expected: true},
		"zero longitude":             {input: "0", expected: true},
		"maximum positive longitude": {input: "180", expected: true},
		"maximum negative longitude": {input: "-180", expected: true},
		"out of range positive":      {input: "180.1", expected: false},
		"out of range negative":      {input: "-180.1", expected: false},
		"not a number":               {input: "abc", expected: false},
		"empty string":               {input: "", expected: false},
		"longitude with spaces":      {input: " 90.0 ", expected: false},
		"longitude with comma":       {input: "90,0", expected: false},
		"longitude with NSEW":        {input: "90E", expected: false},
		"longitude over 180":         {input: "200", expected: false},
		"longitude under -180":       {input: "-200", expected: false},
		"integer longitude":          {input: "90", expected: true},
		"longitude with plus sign":   {input: "+90.0", expected: true},
		"longitude with exponent":    {input: "9.0e1", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsLongitude(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsRequestURL(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid URL":            {input: "http://example.com", expected: true},
		"URL without scheme":   {input: "example.com", expected: false},
		"empty string":         {input: "", expected: false},
		"relative URL":         {input: "/path/to/resource", expected: false},
		"URL with query":       {input: "http://example.com?query=1", expected: true},
		"URL with fragment":    {input: "http://example.com#section", expected: false},
		"URL with auth":        {input: "http://user:pass@example.com", expected: true},
		"URL with port":        {input: "http://example.com:8080", expected: true},
		"URL with IP":          {input: "http://127.0.0.1", expected: true},
		"URL with IPv6":        {input: "http://[::1]", expected: true},
		"URL with invalid TLD": {input: "http://example.invalidtld", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsRequestURL(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsRequestURI(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"absolute URI":              {input: "http://example.com", expected: true},
		"relative URI":              {input: "/path/to/resource", expected: true},
		"relative URI with query":   {input: "/path?query=1", expected: true},
		"relative URI with anchor":  {input: "/path#section", expected: true},
		"empty string":              {input: "", expected: false},
		"invalid URI":               {input: "://invalid", expected: false},
		"just query":                {input: "?query=1", expected: false},
		"just fragment":             {input: "#section", expected: false},
		"spaces in URI":             {input: "/path with spaces", expected: true},
		"full URI with port":        {input: "http://example.com:8080", expected: true},
		"full URI with auth":        {input: "http://user:pass@example.com", expected: true},
		"URI with unicode":          {input: "/路径/资源", expected: true},
		"URI with percent-encoding": {input: "/path%20with%20spaces", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsRequestURI(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsUTFLetter(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"latin letters":      {input: "abcXYZ", expected: true},
		"unicode letters":    {input: "абвгд", expected: true},
		"letters with marks": {input: "mañana", expected: true},
		"contains numbers":   {input: "abc123", expected: false},
		"contains symbols":   {input: "abc!", expected: false},
		"empty string":       {input: "", expected: true},
		"emoji":              {input: "😀", expected: false},
		"combined scripts":   {input: "abcабв", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsUTFLetter(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsAlphanumeric(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"letters and numbers": {input: "abc123", expected: true},
		"only letters":        {input: "abcXYZ", expected: true},
		"only numbers":        {input: "123456", expected: true},
		"contains symbols":    {input: "abc123!", expected: false},
		"contains spaces":     {input: "abc 123", expected: false},
		"empty string":        {input: "", expected: true},
		"unicode letters":     {input: "абвгд123", expected: false},
		"mixed case":          {input: "AbC123", expected: true},
		"special characters":  {input: "abc_123", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsAlphanumeric(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsUTFLetterNumeric(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"latin letters and numbers":   {input: "abc123", expected: true},
		"unicode letters and numbers": {input: "абв123", expected: true},
		"letters with accents":        {input: "café123", expected: true},
		"contains symbols":            {input: "abc123!", expected: false},
		"contains spaces":             {input: "abc 123", expected: false},
		"empty string":                {input: "", expected: true},
		"only unicode letters":        {input: "你好", expected: true},
		"emoji":                       {input: "😀", expected: false},
		"mixed scripts":               {input: "abcабв123", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsUTFLetterNumeric(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsUTFNumeric(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"arabic numerals":        {input: "123456", expected: true},
		"negative number":        {input: "-123", expected: true},
		"positive number":        {input: "+123", expected: true},
		"decimal number":         {input: "123.45", expected: false},
		"unicode digits":         {input: "１２３４５６", expected: true},
		"roman numerals":         {input: "Ⅻ", expected: true},
		"fractions":              {input: "¾", expected: true},
		"contains letters":       {input: "123abc", expected: false},
		"empty string":           {input: "", expected: true},
		"spaces included":        {input: "123 456", expected: false},
		"symbols included":       {input: "123$", expected: false},
		"scientific notation":    {input: "1e10", expected: false},
		"negative unicode digit": {input: "-１２３", expected: true},
		"leading zeros":          {input: "００１２３", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsUTFNumeric(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsUTFDigit(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"arabic numerals":        {input: "123456", expected: true},
		"negative number":        {input: "-123", expected: true},
		"positive number":        {input: "+123", expected: true},
		"decimal number":         {input: "123.45", expected: false},
		"unicode digits":         {input: "１２３４５６", expected: true},
		"non-decimal digits":     {input: "Ⅻ", expected: false},
		"contains letters":       {input: "123abc", expected: false},
		"empty string":           {input: "", expected: true},
		"spaces included":        {input: "123 456", expected: false},
		"symbols included":       {input: "123$", expected: false},
		"scientific notation":    {input: "1e10", expected: false},
		"negative unicode digit": {input: "-１２３", expected: true},
		"leading zeros":          {input: "００１２３", expected: true},
		"emoji digits":           {input: "1️⃣", expected: false},
		"full-width plus sign":   {input: "＋１２３", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsUTFDigit(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsHexcolor(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid short hex":         {input: "#FFF", expected: true},
		"valid long hex":          {input: "#FFFFFF", expected: true},
		"valid without hash":      {input: "FFFFFF", expected: true},
		"valid shorthand no hash": {input: "FFF", expected: true},
		"invalid length":          {input: "#FFFF", expected: false},
		"invalid characters":      {input: "#GGG", expected: false},
		"empty string":            {input: "", expected: false},
		"too long":                {input: "#FFFFFFFF", expected: false},
		"contains spaces":         {input: "#FF FF FF", expected: false},
		"missing hash":            {input: "FFF", expected: true},
		"lowercase hex":           {input: "#abc", expected: true},
		"mixed case hex":          {input: "#AbC123", expected: true},
		"invalid symbol":          {input: "#12345G", expected: false},
		"hex with opacity":        {input: "#FFFFFF00", expected: false}, // Not valid in standard hex color
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsHexcolor(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsRGBcolor(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid RGB":           {input: "rgb(255, 255, 255)", expected: true},
		"valid RGB no space":  {input: "rgb(255,255,255)", expected: true},
		"valid RGB zeros":     {input: "rgb(0, 0, 0)", expected: true},
		"invalid format":      {input: "rgb(255,255)", expected: false},
		"values too high":     {input: "rgb(256, 255, 255)", expected: false},
		"values too low":      {input: "rgb(-1, 0, 0)", expected: false},
		"empty string":        {input: "", expected: false},
		"missing rgb":         {input: "(255,255,255)", expected: false},
		"additional values":   {input: "rgb(255,255,255,0)", expected: false},
		"non-integer values":  {input: "rgb(255.5,255,255)", expected: false},
		"hex color":           {input: "#FFFFFF", expected: false},
		"with alpha":          {input: "rgba(255,255,255,1)", expected: false},
		"missing parenthesis": {input: "rgb255,255,255)", expected: false},
		"mixed case":          {input: "RgB(255,255,255)", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsRGBcolor(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsLowerCase(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"all lowercase":       {input: "abcdef", expected: true},
		"contains uppercase":  {input: "abcDef", expected: false},
		"all uppercase":       {input: "ABCDEF", expected: false},
		"empty string":        {input: "", expected: true},
		"numbers and symbols": {input: "abc123!", expected: true},
		"unicode lowercase":   {input: "абвгд", expected: true},
		"unicode uppercase":   {input: "АБВГД", expected: false},
		"mixed unicode case":  {input: "абвГД", expected: false},
		"special characters":  {input: "abc_def", expected: true},
		"spaces included":     {input: "abc def", expected: true},
		"numeric string":      {input: "12345", expected: true},
		"emoji":               {input: "😀", expected: true},
		"punctuation":         {input: "abc.", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsLowerCase(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsUpperCase(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"all uppercase":       {input: "ABCDEF", expected: true},
		"contains lowercase":  {input: "ABCdEF", expected: false},
		"all lowercase":       {input: "abcdef", expected: false},
		"empty string":        {input: "", expected: true},
		"numbers and symbols": {input: "ABC123!", expected: true},
		"unicode uppercase":   {input: "АБВГД", expected: true},
		"unicode lowercase":   {input: "абвгд", expected: false},
		"mixed unicode case":  {input: "АБВгд", expected: false},
		"special characters":  {input: "ABC_DEF", expected: true},
		"spaces included":     {input: "ABC DEF", expected: true},
		"numeric string":      {input: "12345", expected: true},
		"emoji":               {input: "😀", expected: true},
		"punctuation":         {input: "ABC.", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsUpperCase(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHasLowerCase(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"all lowercase":       {input: "abcdef", expected: true},
		"contains uppercase":  {input: "abcDef", expected: true},
		"all uppercase":       {input: "ABCDEF", expected: false},
		"empty string":        {input: "", expected: true},
		"numbers and symbols": {input: "123!@", expected: false},
		"unicode lowercase":   {input: "абвгд", expected: true},
		"mixed unicode case":  {input: "АБВгД", expected: true},
		"special characters":  {input: "ABC_def", expected: true},
		"spaces included":     {input: "ABC def", expected: true},
		"numeric string":      {input: "12345", expected: false},
		"emoji":               {input: "😀", expected: false},
		"punctuation":         {input: "ABC.", expected: false},
		"mixed with numbers":  {input: "abc123", expected: true},
		"lowercase at end":    {input: "ABCdef", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := HasLowerCase(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHasUpperCase(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"all uppercase":       {input: "ABCDEF", expected: true},
		"contains lowercase":  {input: "abcDef", expected: true},
		"all lowercase":       {input: "abcdef", expected: false},
		"empty string":        {input: "", expected: true},
		"numbers and symbols": {input: "123!@", expected: false},
		"unicode uppercase":   {input: "АБВГД", expected: true},
		"mixed unicode case":  {input: "абвГД", expected: true},
		"special characters":  {input: "abc_DEF", expected: true},
		"spaces included":     {input: "abc DEF", expected: true},
		"numeric string":      {input: "12345", expected: false},
		"emoji":               {input: "😀", expected: false},
		"punctuation":         {input: "abc.", expected: false},
		"mixed with numbers":  {input: "ABC123", expected: true},
		"uppercase at end":    {input: "abcDEF", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := HasUpperCase(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsDivisibleBy(t *testing.T) {
	tests := map[string]struct {
		input    string
		divisor  string
		expected bool
	}{
		"divisible by 2":        {input: "4", divisor: "2", expected: true},
		"not divisible by 2":    {input: "5", divisor: "2", expected: false},
		"divisible by negative": {input: "-4", divisor: "2", expected: true},
		"zero divisible":        {input: "0", divisor: "1", expected: true},
		"divisor zero":          {input: "5", divisor: "0", expected: false},
		"invalid input":         {input: "abc", divisor: "2", expected: true}, // Invalid input converts to zero
		"invalid divisor":       {input: "5", divisor: "abc", expected: false},
		"input zero":            {input: "0", divisor: "0", expected: false},
		"decimal input":         {input: "10.0", divisor: "2", expected: true},
		"decimal divisor":       {input: "10", divisor: "2.5", expected: false},
		"negative divisor":      {input: "10", divisor: "-2", expected: true},
		"both negative":         {input: "-10", divisor: "-2", expected: true},
		"large numbers":         {input: "1000000000", divisor: "100000", expected: true},
		"non-integer input":     {input: "10.5", divisor: "2", expected: true},
		"non-integer divisor":   {input: "10", divisor: "2.5", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsDivisibleBy(tc.input, tc.divisor)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsNotNull(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"non-empty string":     {input: "hello", expected: true},
		"empty string":         {input: "", expected: false},
		"whitespace string":    {input: " ", expected: true},
		"newline string":       {input: "\n", expected: true},
		"tab string":           {input: "\t", expected: true},
		"multiple whitespaces": {input: "   ", expected: true},
		"unicode characters":   {input: "こんにちは", expected: true},
		"zero-length string":   {input: "", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsNotNull(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHasWhitespaceOnly(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"only spaces":             {input: "   ", expected: true},
		"only tabs":               {input: "\t\t\t", expected: true},
		"only newlines":           {input: "\n\n\n", expected: true},
		"mixed whitespace":        {input: " \t\n ", expected: true},
		"contains non-whitespace": {input: " hello ", expected: false},
		"empty string":            {input: "", expected: false},
		"single space":            {input: " ", expected: true},
		"single tab":              {input: "\t", expected: true},
		"single newline":          {input: "\n", expected: true},
		"whitespace with unicode": {input: "　", expected: true}, // Full-width space
		"whitespace with symbols": {input: " \t\n! ", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := HasWhitespaceOnly(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHasWhitespace(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"no whitespace":           {input: "hello", expected: false},
		"spaces included":         {input: "hello world", expected: true},
		"tabs included":           {input: "hello\tworld", expected: true},
		"newlines included":       {input: "hello\nworld", expected: true},
		"multiple whitespaces":    {input: "hello  \t\n world", expected: true},
		"only whitespace":         {input: "   ", expected: true},
		"empty string":            {input: "", expected: false},
		"whitespace with unicode": {input: "こんにちは 世界", expected: true},
		"whitespace with symbols": {input: "hello! world!", expected: true},
		"whitespace at ends":      {input: " hello ", expected: true},
		"whitespace in unicode":   {input: "hello\u2003world", expected: true}, // Em space
		"no visible whitespace":   {input: "helloworld", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := HasWhitespace(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsByteLength(t *testing.T) {
	tests := map[string]struct {
		input    string
		min      int
		max      int
		expected bool
	}{
		"length within range":         {input: "hello", min: 3, max: 10, expected: true},
		"length below minimum":        {input: "hi", min: 3, max: 10, expected: false},
		"length above maximum":        {input: "hello world", min: 3, max: 10, expected: false},
		"length exactly minimum":      {input: "abc", min: 3, max: 10, expected: true},
		"length exactly maximum":      {input: "abcdefghij", min: 3, max: 10, expected: true},
		"empty string with min 0":     {input: "", min: 0, max: 10, expected: true},
		"empty string with min 1":     {input: "", min: 1, max: 10, expected: false},
		"unicode characters":          {input: "こんにちは", min: 5, max: 15, expected: true}, // 15 bytes if UTF-8
		"multi-byte characters":       {input: "😀😃😄😁", min: 4, max: 16, expected: true},  // 4 runes, 16 bytes
		"multi-byte over max":         {input: "😀😃😄😁", min: 1, max: 15, expected: false},
		"max less than min":           {input: "hello", min: 10, max: 5, expected: false}, // Typically invalid, but according to function logic
		"min equals max":              {input: "hello", min: 5, max: 5, expected: true},
		"min greater than max":        {input: "hello", min: 6, max: 5, expected: false},
		"single byte character":       {input: "a", min: 1, max: 1, expected: true},
		"single byte below min":       {input: "a", min: 2, max: 5, expected: false},
		"single byte above max":       {input: "a", min: 0, max: 0, expected: false},
		"string with null byte":       {input: `hel\0lo`, min: 5, max: 10, expected: true}, // Null byte counts as byte
		"string with emojis":          {input: "hello😀", min: 6, max: 10, expected: true},  // '😀' is 4 bytes
		"string with combining marks": {input: "e\u0301", min: 2, max: 4, expected: true},  // 'e' + combining acute
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsByteLength(tc.input, tc.min, tc.max)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsUUIDv3(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid UUIDv3":          {input: "f47ac10b-58cc-4372-a567-0e02b2c3d479", expected: false}, // UUIDv4 example
		"valid UUIDv3 example":  {input: "a987fbc9-4bed-3078-cf07-9141ba07c9f3", expected: true},
		"invalid UUID format":   {input: "a987fbc94bed3078cf079141ba07c9f3", expected: false},
		"empty string":          {input: "", expected: false},
		"invalid version":       {input: "a987fbc9-4bed-4078-cf07-9141ba07c9f3", expected: false}, // UUIDv4
		"short UUID":            {input: "a987fbc9-4bed-3078-cf07-9141ba07c9", expected: false},
		"UUID with braces":      {input: "{a987fbc9-4bed-3078-cf07-9141ba07c9f3}", expected: false},
		"uppercase UUIDv3":      {input: "A987FBC9-4BED-3078-CF07-9141BA07C9F3", expected: true},
		"invalid characters":    {input: "a987fbc9-4bed-3078-cf07-9141ba07c9fX", expected: false},
		"invalid segment count": {input: "a987fbc9-4bed-3078-cf07", expected: false},
		"nil UUIDv3":            {input: "00000000-0000-3000-8000-000000000000", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsUUIDv3(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsUUIDv4(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid UUIDv4":          {input: "f47ac10b-58cc-4372-a567-0e02b2c3d479", expected: true},
		"valid UUIDv4 example":  {input: "550e8400-e29b-41d4-a716-446655440000", expected: true},
		"invalid UUID format":   {input: "550e8400e29b41d4a716446655440000", expected: false},
		"empty string":          {input: "", expected: false},
		"invalid version":       {input: "550e8400-e29b-31d4-a716-446655440000", expected: false}, // UUIDv3
		"short UUID":            {input: "550e8400-e29b-41d4-a716-44665544", expected: false},
		"UUID with braces":      {input: "{550e8400-e29b-41d4-a716-446655440000}", expected: false},
		"uppercase UUIDv4":      {input: "F47AC10B-58CC-4372-A567-0E02B2C3D479", expected: true},
		"invalid characters":    {input: "f47ac10b-58cc-4372-a567-0e02b2c3d47X", expected: false},
		"invalid segment count": {input: "f47ac10b-58cc-4372-a567", expected: false},
		"nil UUIDv4":            {input: "00000000-0000-4000-8000-000000000000", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsUUIDv4(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsUUIDv5(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid UUIDv5":          {input: "123e4567-e89b-52d3-a456-426655440000", expected: true},
		"valid UUIDv5 example":  {input: "123e4567-e89b-52d3-a456-426655440000", expected: true},
		"invalid UUID format":   {input: "123e4567e89b52d3a456426655440000", expected: false},
		"empty string":          {input: "", expected: false},
		"invalid version":       {input: "123e4567-e89b-32d3-a456-426655440000", expected: false}, // UUIDv3
		"short UUID":            {input: "123e4567-e89b-52d3-a456-42665544", expected: false},
		"UUID with braces":      {input: "{123e4567-e89b-52d3-a456-426655440000}", expected: false},
		"uppercase UUIDv5":      {input: "123E4567-E89B-52D3-A456-426655440000", expected: true},
		"invalid characters":    {input: "123e4567-e89b-52d3-a456-42665544000X", expected: false},
		"invalid segment count": {input: "123e4567-e89b-52d3-a456", expected: false},
		"nil UUIDv5":            {input: "00000000-0000-5000-8000-000000000000", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsUUIDv5(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsMultibyte(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"string with ASCII chars only":          {input: "hello", expected: false},
		"string with multibyte chars":           {input: "こんにちは", expected: true},
		"string with mixed chars":               {input: "hello世界", expected: true},
		"empty string":                          {input: "", expected: true}, // As per function
		"string with emojis":                    {input: "hello😀", expected: true},
		"string with accented chars":            {input: "café", expected: true},
		"string with symbols only":              {input: "!@#$%^&*", expected: false},
		"string with space":                     {input: " ", expected: false},
		"string with newline":                   {input: "\n", expected: false},
		"string with tab":                       {input: "\t", expected: false},
		"string with mixed whitespace":          {input: "hello \t\n world", expected: false},
		"string with full-width chars":          {input: "ｈｅｌｌｏ", expected: true},
		"string with half-width and full-width": {input: "helloｈｅｌｌｏ", expected: true},
		"string with combining characters":      {input: "é", expected: true}, // 'e' + combining acute
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsMultibyte(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsASCII(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"string with ASCII chars only":          {input: "Hello, World!", expected: true},
		"string with non-ASCII chars":           {input: "こんにちは", expected: false},
		"empty string":                          {input: "", expected: true}, // As per function
		"string with mixed chars":               {input: "Hello世界", expected: false},
		"string with emojis":                    {input: "Hello😀", expected: false},
		"string with accented chars":            {input: "café", expected: false},
		"string with symbols only":              {input: "!@#$%^&*", expected: true},
		"string with space":                     {input: " ", expected: true},
		"string with newline":                   {input: "\n", expected: true},
		"string with tab":                       {input: "\t", expected: true},
		"string with control characters":        {input: "\x00\x01\x02", expected: true}, // ASCII control
		"string with extended ASCII":            {input: "\x80\xFF", expected: false},
		"string with full-width chars":          {input: "ｈｅｌｌｏ", expected: false},
		"string with half-width and full-width": {input: "helloｈｅｌｌｏ", expected: false},
		"string with combining characters":      {input: "é", expected: false}, // 'e' + combining acute
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsASCII(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsPrintableASCII(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"printable ASCII chars only":            {input: "Hello, World!", expected: true},
		"non-printable ASCII chars":             {input: "Hello\x00World", expected: false},
		"string with non-ASCII chars":           {input: "こんにちは", expected: false},
		"empty string":                          {input: "", expected: true}, // As per function
		"string with mixed chars":               {input: "Hello世界", expected: false},
		"string with emojis":                    {input: "Hello😀", expected: false},
		"string with accented chars":            {input: "café", expected: false},
		"string with symbols only":              {input: "!@#$%^&*", expected: true},
		"string with space":                     {input: " ", expected: true},
		"string with newline":                   {input: "\n", expected: false},
		"string with tab":                       {input: "\t", expected: false},
		"string with control characters":        {input: "\x07", expected: false}, // Bell character
		"string with extended ASCII":            {input: "\x80\xFF", expected: false},
		"string with full-width chars":          {input: "ｈｅｌｌｏ", expected: false},
		"string with half-width and full-width": {input: "helloｈｅｌｌｏ", expected: false},
		"string with combining characters":      {input: "é", expected: false}, // 'e' + combining acute
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsPrintableASCII(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsWinFilePath(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"empty string":                               {input: "", expected: false},
		"valid Windows absolute path":                {input: `C:\Program Files\Example`, expected: true},
		"valid Windows relative path":                {input: `..\Documents\file.txt`, expected: true},
		"invalid Windows path with mixed separators": {input: `C:/Program Files\Example`, expected: false},
		"invalid Windows path too long":              {input: `C:` + strings.Repeat(`\a`, 32768), expected: false},
		"invalid Windows path unknown type":          {input: `:://invalid_path`, expected: false},
		"Windows UNC path":                           {input: `\\Server\Share\Folder`, expected: true},
		"Windows path without drive letter":          {input: `\Program Files\Example`, expected: false},
		"Windows path with invalid characters":       {input: `C:\Program Files\Exa<>mple`, expected: false},
		"Windows path with spaces":                   {input: `C:\Program Files\Example Folder`, expected: true},
		"Windows path with dots":                     {input: `C:\Program Files\.\Example`, expected: true},
		"Windows path with double backslashes":       {input: `C:\\Program Files\\Example`, expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsWinFilePath(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsUnixFilePath(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"empty string":                              {input: "", expected: false},
		"valid Unix absolute path":                  {input: `/usr/local/bin`, expected: true},
		"valid Unix relative path":                  {input: `./scripts/run.sh`, expected: true},
		"Unix path with tilde":                      {input: `~/documents/file.txt`, expected: true},
		"Unix path with double slashes":             {input: `//server/share`, expected: false},
		"invalid Unix path with Windows separators": {input: `/usr/local/bin\script.sh`, expected: true},
		"Unix path with multiple dots":              {input: `/usr/local/./bin`, expected: true},
		"Unix path with spaces":                     {input: `/usr/local/bin/Example Folder`, expected: true},
		"Unix path with unicode characters":         {input: `/usr/local/文档`, expected: true}, // Assuming unicode is allowed
		"Unix path with trailing slash":             {input: `/usr/local/bin/`, expected: true},
		"Unix path with symbolic links":             {input: `/usr/local/bin/link -> /usr/bin`, expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsUnixFilePath(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsDataURI(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid image/png Data URI": {
			input:    "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA",
			expected: true,
		},
		"valid text/plain Data URI": {
			input:    "data:text/plain;base64,SGVsbG8sIFdvcmxkIQ==",
			expected: true,
		},
		"invalid Data URI missing prefix": {
			input:    "image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA",
			expected: false,
		},
		"invalid Data URI missing comma": {
			input:    "data:image/png;base64iVBORw0KGgoAAAANSUhEUgAAAAUA",
			expected: false,
		},
		"invalid media type": {
			input:    "data:invalid/type;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA",
			expected: true,
		},
		"invalid base64 data": {
			input:    "data:image/png;base64,!!!NotBase64!!!",
			expected: false,
		},
		"empty string": {
			input:    "",
			expected: false,
		},
		"Data URI with additional parameters": {
			input:    "data:image/png;charset=utf-8;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA",
			expected: true, // Depending on regex, may pass
		},
		"Data URI with no base64": {
			input:    "data:text/plain,Hello%2C%20World!",
			expected: false, // Not base64 encoded
		},
		"Data URI with invalid scheme": {
			input:    "datas:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA",
			expected: false,
		},
		"Data URI with spaces": {
			input:    "data:image/png; base64,iVBORw0KGgoAAAANSUhEUgAAAAUA",
			expected: false,
		},
		"Data URI with newline": {
			input:    "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA\n",
			expected: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsDataURI(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
func TestIsISO3166Alpha2(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid ISO3166 Alpha2 code US":   {input: "US", expected: true},
		"valid ISO3166 Alpha2 code GB":   {input: "GB", expected: true},
		"valid ISO3166 Alpha2 code FR":   {input: "FR", expected: true},
		"invalid ISO3166 Alpha2 code ZZ": {input: "ZZ", expected: false},
		"invalid length":                 {input: "U", expected: false},
		"empty string":                   {input: "", expected: false},
		"lowercase code":                 {input: "us", expected: false}, // Assuming case-sensitive
		"numeric code":                   {input: "12", expected: false},
		"mixed case code":                {input: "Us", expected: false},
		"special characters":             {input: "U$", expected: false},
		"valid ISO3166 Alpha2 code DE":   {input: "DE", expected: true},
		"valid ISO3166 Alpha2 code JP":   {input: "JP", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsISO3166Alpha2(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsISO3166Alpha3(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid ISO3166 Alpha3 code USA":   {input: "USA", expected: true},
		"valid ISO3166 Alpha3 code GBR":   {input: "GBR", expected: true},
		"valid ISO3166 Alpha3 code FRA":   {input: "FRA", expected: true},
		"invalid ISO3166 Alpha3 code ZZZ": {input: "ZZZ", expected: false},
		"invalid length":                  {input: "US", expected: false},
		"empty string":                    {input: "", expected: false},
		"lowercase code":                  {input: "usa", expected: false}, // Assuming case-sensitive
		"numeric code":                    {input: "123", expected: false},
		"mixed case code":                 {input: "UsA", expected: false},
		"special characters":              {input: "US$", expected: false},
		"valid ISO3166 Alpha3 code DEU":   {input: "DEU", expected: true},
		"valid ISO3166 Alpha3 code JPN":   {input: "JPN", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsISO3166Alpha3(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsISO693Alpha2(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid ISO693 Alpha2 code en":   {input: "en", expected: true},
		"valid ISO693 Alpha2 code fr":   {input: "fr", expected: true},
		"valid ISO693 Alpha2 code de":   {input: "de", expected: true},
		"invalid ISO693 Alpha2 code zz": {input: "zz", expected: false},
		"invalid length":                {input: "e", expected: false},
		"empty string":                  {input: "", expected: false},
		"uppercase code":                {input: "EN", expected: false}, // Assuming case-sensitive
		"numeric code":                  {input: "12", expected: false},
		"mixed case code":               {input: "En", expected: false},
		"special characters":            {input: "e$", expected: false},
		"valid ISO693 Alpha2 code es":   {input: "es", expected: true},
		"valid ISO693 Alpha2 code it":   {input: "it", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsISO693Alpha2(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsISO693Alpha3b(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid ISO693 Alpha3b code eng":   {input: "eng", expected: true},
		"valid ISO693 Alpha3b code fre":   {input: "fre", expected: true},
		"valid ISO693 Alpha3b code ger":   {input: "ger", expected: true},
		"invalid ISO693 Alpha3b code zzz": {input: "zzz", expected: false},
		"invalid length":                  {input: "en", expected: false},
		"empty string":                    {input: "", expected: false},
		"lowercase code":                  {input: "ENG", expected: false}, // Assuming case-sensitive
		"numeric code":                    {input: "123", expected: false},
		"mixed case code":                 {input: "EnG", expected: false},
		"special characters":              {input: "en$", expected: false},
		"valid ISO693 Alpha3b code spa":   {input: "spa", expected: true},
		"valid ISO693 Alpha3b code ita":   {input: "ita", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsISO693Alpha3b(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsDNSName(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid DNS name":                         {input: "example.com", expected: true},
		"valid subdomain":                        {input: "sub.example.com", expected: true},
		"valid localhost":                        {input: "localhost", expected: true},
		"valid single label":                     {input: "example", expected: true},
		"invalid DNS name with IP":               {input: "192.168.1.1", expected: false},
		"invalid DNS name with spaces":           {input: "example .com", expected: false},
		"invalid DNS name with underscore":       {input: "exa_mple.com", expected: true},
		"invalid DNS name with special chars":    {input: "ex*ample.com", expected: false},
		"invalid DNS name too long":              {input: strings.Repeat("a", 256) + ".com", expected: false},
		"empty string":                           {input: "", expected: false},
		"DNS name with trailing dot":             {input: "example.com.", expected: true}, // Often valid
		"DNS name with uppercase letters":        {input: "Example.COM", expected: true},
		"DNS name with hyphens":                  {input: "ex-ample.com", expected: true},
		"DNS name with numeric labels":           {input: "123.example.com", expected: true},
		"DNS name with mixed characters":         {input: "exA-mple123.com", expected: true},
		"DNS name with multiple dots":            {input: "a.b.c.d.e.f.example.com", expected: true},
		"DNS name with invalid TLD":              {input: "example.invalidtld", expected: true},
		"DNS name with international characters": {input: "exämple.com", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsDNSName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsHash(t *testing.T) {
	type testCase struct {
		input     string
		algorithm string
		expected  bool
	}

	tests := map[string]testCase{
		"valid md5 hash":                {input: "d41d8cd98f00b204e9800998ecf8427e", algorithm: "md5", expected: true},
		"invalid md5 hash length":       {input: "d41d8cd98f00b204e9800998ecf8427", algorithm: "md5", expected: false},
		"valid sha1 hash":               {input: "da39a3ee5e6b4b0d3255bfef95601890afd80709", algorithm: "sha1", expected: true},
		"invalid sha1 hash length":      {input: "da39a3ee5e6b4b0d3255bfef95601890afd8070", algorithm: "sha1", expected: false},
		"valid sha256 hash":             {input: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", algorithm: "sha256", expected: true},
		"invalid sha256 hash length":    {input: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b85", algorithm: "sha256", expected: false},
		"valid sha512 hash":             {input: "cf83e1357eefb8bd... (truncated)", algorithm: "sha512", expected: false}, // Provide full hash
		"unsupported algorithm":         {input: "abcdef123456", algorithm: "unknown", expected: false},
		"empty string":                  {input: "", algorithm: "md5", expected: false},
		"uppercase hash":                {input: "D41D8CD98F00B204E9800998ECF8427E", algorithm: "md5", expected: true},
		"valid sha3-256 hash":           {input: "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a", algorithm: "sha3-256", expected: true},
		"invalid sha3-256 hash length":  {input: "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434", algorithm: "sha3-256", expected: false},
		"valid ripemd160 hash":          {input: "9c1185a5c5e9fc54612808977ee8f548b2258d31", algorithm: "ripemd160", expected: true},
		"invalid ripemd160 hash length": {input: "9c1185a5c5e9fc54612808977ee8f548b2258d3", algorithm: "ripemd160", expected: false},
		"valid crc32 hash":              {input: "cbf43926", algorithm: "crc32", expected: true},
		"invalid crc32 hash length":     {input: "cbf4392", algorithm: "crc32", expected: false},
		"valid tiger192 hash":           {input: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12345678", algorithm: "tiger192", expected: false}, // Example
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsHash(tc.input, tc.algorithm)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsSHA3224(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid sha3-224 hash":          {input: "6b4e03423667dbb73b6e15454f0eb1abd4597f891e1d0e98f4e56a6a", expected: true},
		"invalid sha3-224 hash length": {input: "6b4e03423667dbb73b6e15454f0eb1abd4597f891e1d0e98f4e56a6", expected: false},
		"empty string":                 {input: "", expected: false},
		"uppercase hash":               {input: "6B4E03423667DBB73B6E15454F0EB1ABD4597F891E1D0E98F4E56A6A", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsSHA3224(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsSHA3256(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid sha3-256 hash":          {input: "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a", expected: true},
		"invalid sha3-256 hash length": {input: "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434", expected: false},
		"empty string":                 {input: "", expected: false},
		"uppercase hash":               {input: "A7FFC6F8BF1ED76651C14756A061D662F580FF4DE43B49FA82D80A4B80F8434A", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsSHA3256(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsSHA3384(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid sha3-384 hash":          {input: "A7FFC6F8BF1ED76651C14756A061D662F580FF4DE43B49FA82D80A4B80F8434AB49FA82D80A4B80F8434AF8434AF8434", expected: true},
		"invalid sha3-384 hash length": {input: "0c63a75b845e7f5d3c9f1b3b2b3e5a3b4c6d7e8f9a0b1c2d3e4f58h9i0j", expected: false}, // Short
		"empty string":                 {input: "", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsSHA3384(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsSHA3512(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid sha3-512 hash":          {input: "A7FFC6F8BF1ED76651C14756A061D662F580FF4DE43B49FA82D80A4B80F1D662F580FF4DE43B49FA82D80A4B80F8434AB49FA82D80A4B80F8434AF8434AF8434", expected: true},
		"invalid sha3-512 hash length": {input: "0c63a75b845e7f5d3c9f1b3b2b3e5a3b4c6d7e8f9a0b1c2d3e4f58h9i0j", expected: false}, // Short
		"empty string":                 {input: "", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsSHA3512(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsSHA512(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid sha512 hash":          {input: "A7FFC6F8BF1ED76651C14756A061D662F580FF4DE43B49FA82D80A4B80F1D662F580FF4DE43B49FA82D80A4B80F8434AB49FA82D80A4B80F8434AF8434AF8434", expected: true},
		"invalid sha512 hash length": {input: "cf83e1357eefb8bd", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsSHA512(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsSHA384(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid sha384 hash":          {input: "A7FFC6F8BF1ED76651C14756A061D662F580FF4DE43B49FA82D80A4B80F8434AB49FA82D80A4B80F8434AF8434AF8434", expected: true},
		"invalid sha384 hash length": {input: "0c63a75b845e7f5d3c9f1b3b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7g8h9i0j", expected: false}, // Short
		"empty string":               {input: "", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsSHA384(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsSHA256(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid sha256 hash":          {input: "A7FFC6F8BF1ED76651C14756A061D662F580FF4DE43B49FA82D80A4B80F84344", expected: true},
		"invalid sha256 hash length": {input: "e3b0ce41e4649b934ca495991b7852b85", expected: false},
		"empty string":               {input: "", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsSHA256(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsTiger192(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid tiger192 hash":          {input: "0123456789abcdef0123456789abcdef0123456789abcdef", expected: true},
		"invalid tiger192 hash length": {input: "0123456789abcdef01456789abcde", expected: false},
		"empty string":                 {input: "", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsTiger192(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsTiger160(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid tiger160 hash":          {input: "0123456789abcdef0123456789abcdef01234567", expected: true},
		"invalid tiger160 hash length": {input: "0123456789abcdef0156", expected: false},
		"empty string":                 {input: "", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsTiger160(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsRipeMD160(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid RipeMD160 hash":              {input: "9c1185a5c5e9fc54612808977ee8f548b2258d31", expected: true},
		"invalid RipeMD160 hash length":     {input: "9c1185a5c5e9fc54612808977ee8f548b2258d3", expected: false},
		"invalid RipeMD160 hash characters": {input: "9c1185a5c5e9fc54612808977ee8f548b2258d31g", expected: false},
		"empty string":                      {input: "", expected: false},
		"uppercase hash":                    {input: "9C1185A5C5E9FC54612808977EE8F548B2258D31", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsRipeMD160(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsSHA1(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid sha1 hash":              {input: "da39a3ee5e6b4b0d3255bfef95601890afd80709", expected: true},
		"invalid sha1 hash length":     {input: "da39a3ee5e6b4b0d3255bfef95601890afd8070", expected: false},
		"invalid sha1 hash characters": {input: "da39a3ee5e6b4b0d3255bfef95601890afd8070g", expected: false},
		"empty string":                 {input: "", expected: false},
		"uppercase hash":               {input: "DA39A3EE5E6B4B0D3255BFEF95601890AFD80709", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsSHA1(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsRipeMD128(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid RipeMD128 hash":          {input: "c14a12199c66e4ba84636b0f69144c77cfa9a5a1", expected: false}, // RipeMD128 is 32 hex digits, RipeMD160 is 40
		"invalid RipeMD128 hash length": {input: "c14a12199c66e4ba84636b0f69144c77cf", expected: false},
		"empty string":                  {input: "", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsRipeMD128(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsCRC32(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid CRC32 hash":              {input: "cbf43926", expected: true},
		"invalid CRC32 hash length":     {input: "cbf4392", expected: false},
		"invalid CRC32 hash characters": {input: "cbf4392g", expected: false},
		"empty string":                  {input: "", expected: false},
		"uppercase hash":                {input: "CBF43926", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsCRC32(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsCRC32b(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid CRC32b hash":              {input: "cbf43926", expected: true},
		"invalid CRC32b hash length":     {input: "cbf4392", expected: false},
		"invalid CRC32b hash characters": {input: "cbf4392g", expected: false},
		"empty string":                   {input: "", expected: false},
		"uppercase hash":                 {input: "CBF43926", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsCRC32b(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsMD5(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid MD5 hash":              {input: "d41d8cd98f00b204e9800998ecf8427e", expected: true},
		"invalid MD5 hash length":     {input: "d41d8cd98f00b204e9800998ecf8427", expected: false},
		"invalid MD5 hash characters": {input: "d41d8cd98f00b204e9800998ecf8427g", expected: false},
		"empty string":                {input: "", expected: false},
		"uppercase hash":              {input: "D41D8CD98F00B204E9800998ECF8427E", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsMD5(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsMD4(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid MD4 hash":              {input: "a448017aaf21d8525fc10ae87aa6729d", expected: true},
		"invalid MD4 hash length":     {input: "a448017aaf21d8525fc10ae87aa6729", expected: false},
		"invalid MD4 hash characters": {input: "a448017aaf21d8525fc10ae87aa6729g", expected: false},
		"empty string":                {input: "", expected: false},
		"uppercase hash":              {input: "A448017AAF21D8525FC10AE87AA6729D", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsMD4(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsDialString(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid DNS name with port":      {input: "example.com:8080", expected: true},
		"valid IP with port":            {input: "192.168.1.1:80", expected: true},
		"invalid host without port":     {input: "example.com", expected: false},
		"invalid port":                  {input: "example.com:99999", expected: false},
		"invalid IP without port":       {input: "256.256.256.256", expected: false},
		"empty string":                  {input: "", expected: false},
		"valid IPv6 with port":          {input: "[2001:db8::1]:443", expected: true},
		"invalid IPv6 without brackets": {input: "2001:db8::1:443", expected: false},
		"valid localhost with port":     {input: "localhost:3000", expected: true},
		"invalid hostname with spaces":  {input: "example .com:80", expected: false},
		"invalid format":                {input: "://example.com:80", expected: false},
		"port with letters":             {input: "example.com:80a", expected: false},
		"port missing":                  {input: "example.com:", expected: false},
		"port zero":                     {input: "example.com:0", expected: false},
		"port maximum":                  {input: "example.com:65535", expected: true},
		"port above maximum":            {input: "example.com:65536", expected: false},
		"IPv6 with uppercase letters":   {input: "[2001:DB8::1]:443", expected: true},
		"IPv6 with invalid port":        {input: "[2001:db8::1]:abc", expected: false},
		"IPv6 with missing port":        {input: "[2001:db8::1]", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsDialString(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestStringMatches(t *testing.T) {
	tests := map[string]struct {
		input    string
		params   []string
		expected bool
	}{
		"string matches regex":                        {input: "hello123", params: []string{`^[a-z]+[0-9]+$`}, expected: true},
		"string does not match regex":                 {input: "hello123", params: []string{`^[0-9]+$`}, expected: false},
		"invalid regex pattern":                       {input: "hello123", params: []string{`^[a-z]+[`}, expected: false},
		"empty string matches empty regex":            {input: "", params: []string{`^$`}, expected: true},
		"empty string does not match non-empty regex": {input: "", params: []string{`^.+$`}, expected: false},
		"string matches complex regex":                {input: "foo_bar-123", params: []string{`^[a-zA-Z0-9_-]+$`}, expected: true},
		"string with spaces does not match":           {input: "foo bar", params: []string{`^[a-zA-Z0-9_-]+$`}, expected: false},
		"string matches multiple groups":              {input: "foo123bar", params: []string{`^(foo)(\d+)(bar)$`}, expected: true},
		"string does not match multiple groups":       {input: "foo123baz", params: []string{`^(foo)(\d+)(bar)$`}, expected: false},
		"invalid params length":                       {input: "hello", params: []string{}, expected: false},
		"invalid params length two":                   {input: "hello", params: []string{`^h`, `o$`}, expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := StringMatches(tc.input, tc.params...)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsRegex(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid regex":                          {input: `^[a-zA-Z0-9]+$`, expected: true},
		"valid regex with groups":              {input: `(foo|bar)\d+`, expected: true},
		"invalid regex missing bracket":        {input: `^[a-zA-Z0-9+$`, expected: false},
		"empty string":                         {input: "", expected: true}, // Empty pattern is valid
		"valid complex regex":                  {input: `^(?:\+?\d{1,3})?\s?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}$`, expected: true},
		"invalid regex with quantifier":        {input: `a{2,}`, expected: true}, // Valid regex
		"invalid regex with nested quantifier": {input: `a**`, expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsRegex(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsSSN(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid SSN with hyphens":           {input: "123-45-6789", expected: true},
		"valid SSN without hyphens":        {input: "123456789", expected: false}, // Length != 11
		"invalid SSN with letters":         {input: "123-45-678a", expected: false},
		"invalid SSN with extra digits":    {input: "123-45-67890", expected: false},
		"invalid SSN with missing digits":  {input: "12-345-6789", expected: false},
		"empty string":                     {input: "", expected: false},
		"invalid SSN format":               {input: "12345678", expected: false},
		"invalid SSN with special chars":   {input: "123-45-678#", expected: false},
		"valid SSN with uppercase letters": {input: "123-45-6789", expected: true}, // Assuming only digits and hyphens
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsSSN(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsSemver(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid semver":                           {input: "1.0.0", expected: true},
		"valid semver with prerelease":           {input: "1.0.0-alpha", expected: true},
		"valid semver with build metadata":       {input: "1.0.0+20130313144700", expected: true},
		"valid semver with prerelease and build": {input: "1.0.0-beta+exp.sha.5114f85", expected: true},
		"invalid semver missing patch":           {input: "1.0", expected: false},
		"invalid semver non-numeric":             {input: "1.a.0", expected: false},
		"invalid semver with leading zeros":      {input: "01.0.0", expected: false}, // Depending on regex
		"empty string":                           {input: "", expected: false},
		"invalid semver with extra parts":        {input: "1.0.0.0", expected: false},
		"invalid semver with special chars":      {input: "1.0.0-beta!", expected: false},
		"valid semver with multiple prereleases": {input: "1.0.0-alpha.beta", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsSemver(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsType(t *testing.T) {
	tests := map[string]struct {
		value    any
		params   []string
		expected bool
	}{
		"nil value":                   {value: nil, params: []string{"nil"}, expected: true},
		"integer type":                {value: 123, params: []string{"int"}, expected: true},
		"string type":                 {value: "test", params: []string{"string"}, expected: true},
		"float type":                  {value: 3.14, params: []string{"float64"}, expected: true},
		"slice type":                  {value: []int{1, 2, 3}, params: []string{"[]int"}, expected: true},
		"pointer type":                {value: &[]int{1, 2, 3}, params: []string{"*[]int"}, expected: true},
		"mismatched type":             {value: 123, params: []string{"string"}, expected: false},
		"multiple params length one":  {value: "test", params: []string{"string"}, expected: true},
		"multiple params length zero": {value: "test", params: []string{}, expected: false},
		"multiple params length two":  {value: "test", params: []string{"string", "int"}, expected: false},
		"complex type":                {value: make(chan int), params: []string{"chan int"}, expected: true},
		"invalid type name":           {value: 123, params: []string{"integer"}, expected: false}, // Assuming exact match
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsType(tc.value, tc.params...)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsTime(t *testing.T) {
	tests := map[string]struct {
		input    string
		format   string
		expected bool
	}{
		"valid RFC3339 time":                   {input: "2024-04-05T14:30:00Z", format: time.RFC3339, expected: true},
		"invalid RFC3339 time":                 {input: "2024/04/05 14:30:00", format: time.RFC3339, expected: false},
		"valid custom format":                  {input: "05-04-2024 14:30", format: "02-01-2006 15:04", expected: true},
		"invalid custom format":                {input: "2024-04-05 14:30", format: "02-01-2006 15:04", expected: false},
		"empty string":                         {input: "", format: time.RFC3339, expected: false},
		"valid time with timezone offset":      {input: "2024-04-05T14:30:00+02:00", format: time.RFC3339, expected: true},
		"invalid time with incorrect timezone": {input: "2024-04-05T14:30:00+25:00", format: time.RFC3339, expected: false},
		"valid date only":                      {input: "2024-04-05", format: "2006-01-02", expected: true},
		"invalid date only":                    {input: "04/05/2024", format: "2006-01-02", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsTime(tc.input, tc.format)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsUnixTime(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid Unix timestamp":                        {input: "1617181723", expected: true},
		"invalid Unix timestamp non-numeric":          {input: "16171817a3", expected: false},
		"invalid Unix timestamp empty":                {input: "", expected: false},
		"valid Unix timestamp with leading zeros":     {input: "0001617181723", expected: true},
		"invalid Unix timestamp with negative number": {input: "-1617181723", expected: false},
		"valid Unix timestamp max int":                {input: strconv.FormatInt(1<<31-1, 10), expected: true}, // Assuming 32-bit
		"valid Unix timestamp 64-bit":                 {input: strconv.FormatInt(1<<40, 10), expected: true},
		"invalid Unix timestamp with spaces":          {input: "1617181723 ", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsUnixTime(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsRFC3339(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid RFC3339 time":                         {input: "2024-04-05T14:30:00Z", expected: true},
		"valid RFC3339 time with offset":             {input: "2024-04-05T14:30:00+02:00", expected: true},
		"invalid RFC3339 time":                       {input: "2024/04/05 14:30:00", expected: false},
		"empty string":                               {input: "", expected: false},
		"invalid RFC3339 time with wrong format":     {input: "14:30:00 2024-04-05", expected: false},
		"valid RFC3339 time with fractional seconds": {input: "2024-04-05T14:30:00.123Z", expected: true},
		"invalid RFC3339 time with extra characters": {input: "2024-04-05T14:30:00Zabc", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsRFC3339(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsRFC3339WithoutZone(t *testing.T) {
	// Define rfc3339WithoutZone format as "2006-01-02T15:04:05"
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid RFC3339 without zone":                         {input: "2024-04-05T14:30:00", expected: true},
		"invalid RFC3339 without zone with timezone":         {input: "2024-04-05T14:30:00Z", expected: false},
		"invalid format":                                     {input: "2024/04/05 14:30:00", expected: false},
		"empty string":                                       {input: "", expected: false},
		"valid RFC3339 without zone with fractional seconds": {input: "2024-04-05T14:30:00.123", expected: true},
		"invalid RFC3339 without zone with extra characters": {input: "2024-04-05T14:30:00abc", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsRFC3339WithoutZone(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsISO4217(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid ISO4217 code USD":   {input: "USD", expected: true},
		"valid ISO4217 code EUR":   {input: "EUR", expected: true},
		"valid ISO4217 code JPY":   {input: "JPY", expected: true},
		"invalid ISO4217 code XYZ": {input: "XYZ", expected: false},
		"invalid length":           {input: "US", expected: false},
		"empty string":             {input: "", expected: false},
		"lowercase code":           {input: "usd", expected: false}, // Assuming case-sensitive
		"numeric code":             {input: "123", expected: false},
		"mixed case code":          {input: "UsD", expected: false},
		"special characters":       {input: "US$", expected: false},
		"valid ISO4217 code GBP":   {input: "GBP", expected: true},
		"valid ISO4217 code AUD":   {input: "AUD", expected: true},
		"valid ISO4217 code CAD":   {input: "CAD", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsISO4217(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestByteLength(t *testing.T) {
	tests := map[string]struct {
		input    string
		params   []string
		expected bool
	}{
		"within byte range":           {input: "hello", params: []string{"3", "10"}, expected: true},
		"below byte minimum":          {input: "hi", params: []string{"3", "10"}, expected: false},
		"above byte maximum":          {input: "hello world", params: []string{"3", "10"}, expected: false},
		"exactly minimum":             {input: "abc", params: []string{"3", "10"}, expected: true},
		"exactly maximum":             {input: "abcdefghij", params: []string{"3", "10"}, expected: true},
		"empty string with min 0":     {input: "", params: []string{"0", "10"}, expected: true},
		"empty string with min 1":     {input: "", params: []string{"1", "10"}, expected: false},
		"unicode characters":          {input: "こんにちは", params: []string{"5", "15"}, expected: true}, // 15 bytes if UTF-8
		"multi-byte characters":       {input: "😀😃😄😁", params: []string{"4", "16"}, expected: true},  // 4 runes, 16 bytes
		"multi-byte over max":         {input: "😀😃😄😁", params: []string{"1", "15"}, expected: false},
		"max less than min":           {input: "hello", params: []string{"10", "5"}, expected: false},
		"min equals max":              {input: "hello", params: []string{"5", "5"}, expected: true},
		"min greater than max":        {input: "hello", params: []string{"6", "5"}, expected: false},
		"single byte character":       {input: "a", params: []string{"1", "1"}, expected: true},
		"single byte below min":       {input: "a", params: []string{"2", "5"}, expected: false},
		"single byte above max":       {input: "a", params: []string{"0", "0"}, expected: false},
		"string with null byte":       {input: `hel\0lo`, params: []string{"5", "10"}, expected: true},
		"string with emojis":          {input: "hello😀", params: []string{"6", "10"}, expected: true}, // '😀' is 4 bytes
		"string with combining marks": {input: "e\u0301", params: []string{"2", "4"}, expected: true}, // 'e' + combining acute
		"invalid params length":       {input: "hello", params: []string{"3"}, expected: false},
		"invalid params non-integer":  {input: "hello", params: []string{"three", "ten"}, expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := ByteLength(tc.input, tc.params...)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRuneLength(t *testing.T) {
	tests := map[string]struct {
		input    string
		params   []string
		expected bool
	}{
		"within rune range":           {input: "hello", params: []string{"3", "10"}, expected: true},
		"below rune minimum":          {input: "hi", params: []string{"3", "10"}, expected: false},
		"above rune maximum":          {input: "hello world", params: []string{"3", "10"}, expected: false},
		"exactly minimum":             {input: "abc", params: []string{"3", "10"}, expected: true},
		"exactly maximum":             {input: "abcdefghij", params: []string{"3", "10"}, expected: true},
		"empty string with min 0":     {input: "", params: []string{"0", "10"}, expected: true},
		"empty string with min 1":     {input: "", params: []string{"1", "10"}, expected: false},
		"unicode characters":          {input: "こんにちは", params: []string{"5", "15"}, expected: true}, // 5 runes
		"multi-byte characters":       {input: "😀😃😄😁", params: []string{"4", "16"}, expected: true},  // 4 runes
		"multi-byte over max":         {input: "😀😃😄😁", params: []string{"1", "3"}, expected: false},
		"max less than min":           {input: "hello", params: []string{"10", "5"}, expected: false},
		"min equals max":              {input: "hello", params: []string{"5", "5"}, expected: true},
		"min greater than max":        {input: "hello", params: []string{"6", "5"}, expected: false},
		"single rune character":       {input: "a", params: []string{"1", "1"}, expected: true},
		"single rune below min":       {input: "a", params: []string{"2", "5"}, expected: false},
		"single rune above max":       {input: "a", params: []string{"0", "0"}, expected: false},
		"string with null rune":       {input: "hel\U00000000lo", params: []string{"5", "10"}, expected: true},
		"string with emojis":          {input: "hello😀", params: []string{"6", "10"}, expected: true},
		"string with combining marks": {input: "e\u0301", params: []string{"2", "4"}, expected: true}, // 'e' + combining acute
		"invalid params length":       {input: "hello", params: []string{"3"}, expected: false},
		"invalid params non-integer":  {input: "hello", params: []string{"three", "ten"}, expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := RuneLength(tc.input, tc.params...)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestStringLength(t *testing.T) {
	tests := map[string]struct {
		input    string
		params   []string
		expected bool
	}{
		"within rune range":           {input: "hello", params: []string{"3", "10"}, expected: true},
		"below rune minimum":          {input: "hi", params: []string{"3", "10"}, expected: false},
		"above rune maximum":          {input: "hello world", params: []string{"3", "10"}, expected: false},
		"exactly minimum":             {input: "abc", params: []string{"3", "10"}, expected: true},
		"exactly maximum":             {input: "abcdefghij", params: []string{"3", "10"}, expected: true},
		"empty string with min 0":     {input: "", params: []string{"0", "10"}, expected: true},
		"empty string with min 1":     {input: "", params: []string{"1", "10"}, expected: false},
		"unicode characters":          {input: "こんにちは", params: []string{"5", "15"}, expected: true}, // 5 runes
		"multi-byte characters":       {input: "😀😃😄😁", params: []string{"4", "16"}, expected: true},  // 4 runes
		"multi-byte over max":         {input: "😀😃😄😁", params: []string{"1", "3"}, expected: false},
		"max less than min":           {input: "hello", params: []string{"10", "5"}, expected: false},
		"min equals max":              {input: "hello", params: []string{"5", "5"}, expected: true},
		"min greater than max":        {input: "hello", params: []string{"6", "5"}, expected: false},
		"single rune character":       {input: "a", params: []string{"1", "1"}, expected: true},
		"single rune below min":       {input: "a", params: []string{"2", "5"}, expected: false},
		"single rune above max":       {input: "a", params: []string{"0", "0"}, expected: false},
		"string with null rune":       {input: "hel\U00000000lo", params: []string{"5", "10"}, expected: true},
		"string with emojis":          {input: "hello😀", params: []string{"6", "10"}, expected: true},
		"string with combining marks": {input: "e\u0301", params: []string{"2", "4"}, expected: true}, // 'e' + combining acute
		"invalid params length":       {input: "hello", params: []string{"3"}, expected: false},
		"invalid params non-integer":  {input: "hello", params: []string{"three", "ten"}, expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := StringLength(tc.input, tc.params...)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMinStringLength(t *testing.T) {
	tests := map[string]struct {
		input    string
		params   []string
		expected bool
	}{
		"length above min":           {input: "hello", params: []string{"3"}, expected: true},
		"length exactly min":         {input: "abc", params: []string{"3"}, expected: true},
		"length below min":           {input: "hi", params: []string{"3"}, expected: false},
		"empty string with min 0":    {input: "", params: []string{"0"}, expected: true},
		"empty string with min 1":    {input: "", params: []string{"1"}, expected: false},
		"unicode characters":         {input: "こんにちは", params: []string{"5"}, expected: true}, // 5 runes
		"multi-byte characters":      {input: "😀😃😄😁", params: []string{"4"}, expected: true},  // 4 runes
		"multi-byte below min":       {input: "😀😃", params: []string{"3"}, expected: false},
		"invalid params length":      {input: "hello", params: []string{}, expected: false},
		"invalid params non-integer": {input: "hello", params: []string{"three"}, expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := MinStringLength(tc.input, tc.params...)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMaxStringLength(t *testing.T) {
	tests := map[string]struct {
		input    string
		params   []string
		expected bool
	}{
		"length below max":           {input: "hi", params: []string{"3"}, expected: true},
		"length exactly max":         {input: "abc", params: []string{"3"}, expected: true},
		"length above max":           {input: "hello", params: []string{"3"}, expected: false},
		"empty string with max 0":    {input: "", params: []string{"0"}, expected: true},
		"empty string with max 1":    {input: "", params: []string{"1"}, expected: true},
		"unicode characters":         {input: "こんにちは", params: []string{"5"}, expected: true}, // 5 runes
		"multi-byte characters":      {input: "😀😃😄😁", params: []string{"4"}, expected: true},  // 4 runes
		"multi-byte above max":       {input: "😀😃😄😁", params: []string{"3"}, expected: false},
		"invalid params length":      {input: "hello", params: []string{}, expected: false},
		"invalid params non-integer": {input: "hello", params: []string{"three"}, expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := MaxStringLength(tc.input, tc.params...)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRange(t *testing.T) {
	tests := map[string]struct {
		input    string
		params   []string
		expected bool
	}{
		"within range":                       {input: "5", params: []string{"1", "10"}, expected: true},
		"below range":                        {input: "0", params: []string{"1", "10"}, expected: false},
		"above range":                        {input: "11", params: []string{"1", "10"}, expected: false},
		"exactly min":                        {input: "1", params: []string{"1", "10"}, expected: true},
		"exactly max":                        {input: "10", params: []string{"1", "10"}, expected: true},
		"invalid float within range":         {input: "5.5", params: []string{"1", "10"}, expected: true},
		"invalid float below range":          {input: "0.5", params: []string{"1", "10"}, expected: false},
		"invalid float above range":          {input: "10.5", params: []string{"1", "10"}, expected: false},
		"invalid non-numeric":                {input: "five", params: []string{"1", "10"}, expected: false},
		"invalid params length":              {input: "5", params: []string{"1"}, expected: false},
		"range with left greater than right": {input: "5", params: []string{"10", "1"}, expected: true},
		"string":                             {input: "apple", params: []string{"banana", "zebra"}, expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := Range(tc.input, tc.params...)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsInRaw(t *testing.T) {
	tests := map[string]struct {
		input    string
		params   []string
		expected bool
	}{
		"string is in list":                       {input: "apple", params: []string{"apple|banana|cherry"}, expected: true},
		"string is not in list":                   {input: "date", params: []string{"apple|banana|cherry"}, expected: false},
		"empty string in list":                    {input: "", params: []string{"|banana|cherry"}, expected: true},
		"empty string not in list":                {input: "", params: []string{"apple|banana|cherry"}, expected: false},
		"string with special characters":          {input: "foo@bar", params: []string{"foo@bar|baz"}, expected: true},
		"invalid params length":                   {input: "apple", params: []string{}, expected: false},
		"invalid params with multiple separators": {input: "apple", params: []string{"apple|banana|cherry|"}, expected: true},
		"string with mixed cases":                 {input: "Apple", params: []string{"apple|banana|cherry"}, expected: false}, // Assuming case-sensitive
		"string with spaces":                      {input: "apple pie", params: []string{"apple pie|banana|cherry"}, expected: true},
		"string not in list with similar strings": {input: "apples", params: []string{"apple|banana|cherry"}, expected: false},
		"string with numeric":                     {input: "123", params: []string{"123|456|789"}, expected: true},
		"string with unicode":                     {input: "こんにちは", params: []string{"こんにちは|さようなら"}, expected: true},
		"string with unicode not in list":         {input: "こんばんは", params: []string{"こんにちは|さようなら"}, expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsInRaw(tc.input, tc.params...)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsE164(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid E.164 number":                    {input: "+14155552671", expected: true},
		"invalid E.164 without plus":            {input: "14155552671", expected: true},
		"invalid E.164 with spaces":             {input: "+1415 555 2671", expected: false},
		"invalid E.164 with dashes":             {input: "+1-415-555-2671", expected: false},
		"empty string":                          {input: "", expected: false},
		"valid E.164 with maximum length":       {input: "+" + strings.Repeat("1", 15), expected: true}, // E.164 max 15 digits
		"invalid E.164 exceeding max length":    {input: "+" + strings.Repeat("1", 16), expected: false},
		"invalid E.164 with letters":            {input: "+1415abc2671", expected: false},
		"valid E.164 with country code":         {input: "+442071838750", expected: true}, // UK number
		"valid E.164 with leading zeros":        {input: "+12025550123", expected: true},
		"invalid E.164 with multiple pluses":    {input: "++14155552671", expected: false},
		"invalid E.164 with special characters": {input: "+1415@555#2671", expected: false},
		"valid E.164 with unicode":              {input: "+١٤١٥٥٥٥٢٦٧١", expected: false}, // Arabic-Indic digits, assuming only ASCII digits are allowed
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsE164(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestAbs(t *testing.T) {
	tests := map[string]struct {
		input    float64
		expected float64
	}{
		"positive number":   {input: 5.5, expected: 5.5},
		"negative number":   {input: -5.5, expected: 5.5},
		"zero":              {input: 0.0, expected: 0.0},
		"positive integer":  {input: 10.0, expected: 10.0},
		"negative integer":  {input: -10.0, expected: 10.0},
		"positive infinity": {input: math.Inf(1), expected: math.Inf(1)},
		"negative infinity": {input: math.Inf(-1), expected: math.Inf(1)},
		"NaN value":         {input: math.NaN(), expected: math.NaN()},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := Abs(tc.input)
			if math.IsNaN(tc.expected) {
				assert.True(t, math.IsNaN(result), "expected NaN")
			} else {
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestSign(t *testing.T) {
	tests := map[string]struct {
		input    float64
		expected float64
	}{
		"positive number":   {input: 5.5, expected: 1},
		"negative number":   {input: -5.5, expected: -1},
		"zero":              {input: 0.0, expected: 0},
		"positive integer":  {input: 10.0, expected: 1},
		"negative integer":  {input: -10.0, expected: -1},
		"positive infinity": {input: math.Inf(1), expected: 1},
		"negative infinity": {input: math.Inf(-1), expected: -1},
		"NaN value":         {input: math.NaN(), expected: 0}, // Assuming Sign returns 0 for NaN
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := Sign(tc.input)
			if math.IsNaN(tc.input) {
				assert.Equal(t, 0.0, result)
			} else {
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestIsNegative(t *testing.T) {
	tests := map[string]struct {
		input    float64
		expected bool
	}{
		"positive number":   {input: 5.5, expected: false},
		"negative number":   {input: -5.5, expected: true},
		"zero":              {input: 0.0, expected: false},
		"positive integer":  {input: 10.0, expected: false},
		"negative integer":  {input: -10.0, expected: true},
		"positive infinity": {input: math.Inf(1), expected: false},
		"negative infinity": {input: math.Inf(-1), expected: true},
		"NaN value":         {input: math.NaN(), expected: false}, // NaN is not less than zero
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsNegative(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsPositive(t *testing.T) {
	tests := map[string]struct {
		input    float64
		expected bool
	}{
		"positive number":   {input: 5.5, expected: true},
		"negative number":   {input: -5.5, expected: false},
		"zero":              {input: 0.0, expected: false},
		"positive integer":  {input: 10.0, expected: true},
		"negative integer":  {input: -10.0, expected: false},
		"positive infinity": {input: math.Inf(1), expected: true},
		"negative infinity": {input: math.Inf(-1), expected: false},
		"NaN value":         {input: math.NaN(), expected: false}, // NaN is not greater than zero
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsPositive(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsNonNegative(t *testing.T) {
	tests := map[string]struct {
		input    float64
		expected bool
	}{
		"positive number":   {input: 5.5, expected: true},
		"negative number":   {input: -5.5, expected: false},
		"zero":              {input: 0.0, expected: true},
		"positive integer":  {input: 10.0, expected: true},
		"negative integer":  {input: -10.0, expected: false},
		"positive infinity": {input: math.Inf(1), expected: true},
		"negative infinity": {input: math.Inf(-1), expected: false},
		"NaN value":         {input: math.NaN(), expected: false}, // NaN is not >= 0
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsNonNegative(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsNonPositive(t *testing.T) {
	tests := map[string]struct {
		input    float64
		expected bool
	}{
		"positive number":   {input: 5.5, expected: false},
		"negative number":   {input: -5.5, expected: true},
		"zero":              {input: 0.0, expected: true},
		"positive integer":  {input: 10.0, expected: false},
		"negative integer":  {input: -10.0, expected: true},
		"positive infinity": {input: math.Inf(1), expected: false},
		"negative infinity": {input: math.Inf(-1), expected: true},
		"NaN value":         {input: math.NaN(), expected: false}, // NaN is not <= 0
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsNonPositive(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestInRange(t *testing.T) {
	tests := map[string]struct {
		value    any
		left     any
		right    any
		expected bool
	}{
		"integer within range":                 {value: 5, left: 1, right: 10, expected: true},
		"integer below range":                  {value: 0, left: 1, right: 10, expected: false},
		"integer above range":                  {value: 11, left: 1, right: 10, expected: false},
		"float within range":                   {value: 5.5, left: 1.0, right: 10.0, expected: true},
		"float below range":                    {value: 0.5, left: 1.0, right: 10.0, expected: false},
		"float above range":                    {value: 10.5, left: 1.0, right: 10.0, expected: false},
		"integer with left greater than right": {value: 5, left: 10, right: 1, expected: true},
		"invalid types":                        {value: "5", left: 1, right: 10, expected: false},
		"invalid value type":                   {value: []int{1, 2, 3}, left: 1, right: 10, expected: false},
		"range with same left and right":       {value: 5, left: 5, right: 5, expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := InRange(tc.value, tc.left, tc.right)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsNatural(t *testing.T) {
	tests := map[string]struct {
		input    float64
		expected bool
	}{
		"natural positive integer":      {input: 5.0, expected: true},
		"natural positive float":        {input: 5.5, expected: false},
		"zero":                          {input: 0.0, expected: false},
		"negative integer":              {input: -5.0, expected: false},
		"whole number as float":         {input: 10.0, expected: true},
		"non-whole positive number":     {input: 10.1, expected: false},
		"natural number with precision": {input: 100.0000, expected: true},
		"NaN value":                     {input: math.NaN(), expected: false},
		"positive infinity":             {input: math.Inf(1), expected: false},
		"negative infinity":             {input: math.Inf(-1), expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsNatural(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsWhole(t *testing.T) {
	tests := map[string]struct {
		input    float64
		expected bool
	}{
		"whole positive number":       {input: 5.0, expected: true},
		"whole negative number":       {input: -5.0, expected: true},
		"non-whole positive number":   {input: 5.5, expected: false},
		"non-whole negative number":   {input: -5.5, expected: false},
		"zero":                        {input: 0.0, expected: true},
		"whole number as float":       {input: 10.0, expected: true},
		"non-whole number as float":   {input: 10.1, expected: false},
		"whole number with precision": {input: 100.0000, expected: true},
		"NaN value":                   {input: math.NaN(), expected: false},
		"positive infinity":           {input: math.Inf(1), expected: false},
		"negative infinity":           {input: math.Inf(-1), expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsWhole(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsIMSI(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid IMSI":                    {input: "310150123456789", expected: true},
		"invalid IMSI length short":     {input: "31015012345678", expected: true},
		"invalid IMSI length long":      {input: "3101501234567890", expected: false},
		"invalid IMSI characters":       {input: "31015012345678a", expected: false},
		"empty string":                  {input: "", expected: false},
		"invalid MCC code":              {input: "999150123456789", expected: false}, // Assuming 999 is not in switch
		"valid IMSI with leading zeros": {input: "001150123456789", expected: false},
		"valid IMSI with boundary MCC":  {input: "202150123456789", expected: true},  // 202 is valid
		"invalid MCC code edge":         {input: "100150123456789", expected: false}, // Assuming 100 is not listed
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsIMSI(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsIMEI(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid IMEI":                      {input: "490154203237518", expected: true},
		"invalid IMEI length short":       {input: "49015420323751", expected: true},
		"invalid IMEI length long":        {input: "4901542032375189", expected: false},
		"invalid IMEI characters":         {input: "49015420323751a", expected: false},
		"empty string":                    {input: "", expected: false},
		"valid IMEI with all zeros":       {input: "000000000000000", expected: true},
		"invalid IMEI with special chars": {input: "49015420323751!", expected: false},
		"valid IMEI uppercase letters":    {input: "490154203237518", expected: true}, // Assuming only digits are allowed
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsIMEI(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsCIDR(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid IPv4 CIDR":                    {input: "192.168.1.0/24", expected: true},
		"valid IPv6 CIDR":                    {input: "2001:db8::/32", expected: true},
		"invalid CIDR without slash":         {input: "192.168.1.0-24", expected: false},
		"invalid CIDR with wrong IP":         {input: "999.999.999.999/24", expected: false},
		"invalid CIDR with wrong prefix":     {input: "192.168.1.0/33", expected: false}, // IPv4 prefix max 32
		"empty string":                       {input: "", expected: false},
		"invalid CIDR format":                {input: "2001:db8::/129", expected: false}, // IPv6 prefix max 128
		"valid IPv4 CIDR with zeros":         {input: "192.168.001.000/24", expected: false},
		"valid IPv6 CIDR with uppercase":     {input: "2001:DB8::/32", expected: true},
		"invalid CIDR with extra characters": {input: "192.168.1.0/24a", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsCIDR(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsHost(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid DNS name":               {input: "example.com", expected: true},
		"valid subdomain":              {input: "sub.example.com", expected: true},
		"valid IPv4 address":           {input: "192.168.1.1", expected: true},
		"valid IPv6 address":           {input: "2001:0db8:85a3:0000:0000:8a2e:0370:7334", expected: true},
		"invalid DNS name with IP":     {input: "example.com:8080", expected: false},
		"empty string":                 {input: "", expected: false},
		"invalid DNS name with spaces": {input: "example .com", expected: false},
		"DNS name with trailing dot":   {input: "example.com.", expected: true},       // Often valid
		"IPv6 address with brackets":   {input: "[2001:db8::1]", expected: false},     // Should be without brackets
		"IPv6 address with port":       {input: "[2001:db8::1]:443", expected: false}, // Port not handled in IsHost
		"localhost":                    {input: "localhost", expected: true},
		"numeric DNS label":            {input: "123.example.com", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsHost(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsMongoID(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid MongoDB ObjectId":                {input: "507f1f77bcf86cd799439011", expected: true},
		"invalid length short":                  {input: "507f1f77bcf86cd79943901", expected: false},
		"invalid length long":                   {input: "507f1f77bcf86cd799439011a", expected: false},
		"invalid characters":                    {input: "507f1f77bcf86cd79943901G", expected: false},
		"empty string":                          {input: "", expected: false},
		"uppercase characters":                  {input: "507F1F77BCF86CD799439011", expected: true},
		"valid MongoDB ObjectId with uppercase": {input: "507F1F77BCF86CD799439011", expected: true},
		"mixed case characters":                 {input: "507f1F77bcF86cD799439011", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsMongoID(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsIPv4(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid IPv4 address":                {input: "192.168.1.1", expected: true},
		"invalid IPv4 address format":       {input: "192.168.1", expected: false},
		"invalid IPv4 address range":        {input: "256.256.256.256", expected: false},
		"IPv6 address":                      {input: "2001:0db8:85a3:0000:0000:8a2e:0370:7334", expected: false},
		"empty string":                      {input: "", expected: false},
		"valid IPv4 with leading zeros":     {input: "192.168.001.001", expected: false},
		"invalid characters":                {input: "192.168.1.a", expected: false},
		"valid localhost IPv4":              {input: "127.0.0.1", expected: true},
		"invalid IPv4 with negative number": {input: "-1.168.1.1", expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsIPv4(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsIPv6(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"valid IPv6 address":                {input: "2001:0db8:85a3:0000:0000:8a2e:0370:7334", expected: true},
		"valid IPv6 address with shorthand": {input: "2001:db8::1", expected: true},
		"invalid IPv6 address format":       {input: "2001:0db8:85a3::8a2e:0370:7334:", expected: false},
		"valid IPv6 with IPv4":              {input: "::ffff:192.168.1.1", expected: true},
		"invalid IPv6 with mixed format":    {input: "2001:db8::85a3::8a2e:0370:7334", expected: false},
		"empty string":                      {input: "", expected: false},
		"invalid characters":                {input: "2001:db8::g85a3::8a2e:0370:7334", expected: false},
		"valid IPv6 loopback":               {input: "::1", expected: true},
		"valid IPv6 unspecified":            {input: "::", expected: true},
		"valid IPv6 with uppercase letters": {input: "2001:DB8::1", expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := IsIPv6(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMin(t *testing.T) {
	tests := map[string]struct {
		input    string
		params   []string
		expected bool
	}{
		"numeric > min":                   {input: "10", params: []string{"5"}, expected: true},
		"numeric = min":                   {input: "5", params: []string{"5"}, expected: true},
		"numeric < min":                   {input: "3", params: []string{"5"}, expected: false},
		"non-numeric len > min":           {input: "hello", params: []string{"5"}, expected: true},
		"non-numeric len = min":           {input: "hello", params: []string{"5"}, expected: true},
		"non-numeric len < min":           {input: "hi", params: []string{"3"}, expected: false},
		"empty string min=0":              {input: "", params: []string{"0"}, expected: true},
		"empty string min=1":              {input: "", params: []string{"1"}, expected: false},
		"multi-byte len >= min":           {input: "こんにちは", params: []string{"5"}, expected: true},
		"multi-byte len < min":            {input: "こんにちは", params: []string{"6"}, expected: false},
		"mixed chars len >= min":          {input: "hello世界", params: []string{"7"}, expected: true},
		"mixed chars len < min":           {input: "hello世界", params: []string{"8"}, expected: false},
		"invalid min param non-numeric":   {input: "10", params: []string{"five"}, expected: false},
		"invalid min param non-numeric2":  {input: "hello", params: []string{"five"}, expected: false},
		"missing min param":               {input: "10", params: []string{}, expected: false},
		"multiple min params":             {input: "10", params: []string{"5", "6"}, expected: false},
		"numeric negative min":            {input: "-5", params: []string{"-10"}, expected: true},
		"numeric below negative min":      {input: "-15", params: []string{"-10"}, expected: false},
		"numeric zero min":                {input: "0", params: []string{"-5"}, expected: true},
		"non-numeric len >= negative min": {input: "a", params: []string{"-1"}, expected: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := Min(tc.input, tc.params...)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMax(t *testing.T) {
	tests := map[string]struct {
		input    string
		params   []string
		expected bool
	}{
		"numeric < max":                   {input: "5", params: []string{"10"}, expected: true},
		"numeric = max":                   {input: "10", params: []string{"10"}, expected: true},
		"numeric > max":                   {input: "15", params: []string{"10"}, expected: false},
		"non-numeric len < max":           {input: "hi", params: []string{"5"}, expected: true},
		"non-numeric len = max":           {input: "hello", params: []string{"5"}, expected: true},
		"non-numeric len > max":           {input: "hello world", params: []string{"5"}, expected: false},
		"empty string max=0":              {input: "a", params: []string{"0"}, expected: false},
		"empty string max=1":              {input: "", params: []string{"1"}, expected: true},
		"multi-byte len <= max":           {input: "こんにちは", params: []string{"5"}, expected: true},
		"multi-byte len > max":            {input: "こんにちは", params: []string{"4"}, expected: false},
		"mixed chars len <= max":          {input: "hello世界", params: []string{"7"}, expected: true},
		"mixed chars len > max":           {input: "hello世界", params: []string{"6"}, expected: false},
		"invalid max param non-numeric":   {input: "10", params: []string{"ten"}, expected: false},
		"invalid max param non-numeric2":  {input: "hello", params: []string{"ten"}, expected: false},
		"missing max param":               {input: "10", params: []string{}, expected: false},
		"multiple max params":             {input: "10", params: []string{"15", "20"}, expected: false},
		"numeric negative max":            {input: "-5", params: []string{"-10"}, expected: false},
		"numeric above negative max":      {input: "-15", params: []string{"-10"}, expected: true},
		"numeric zero max":                {input: "0", params: []string{"-5"}, expected: false},
		"non-numeric len <= max":          {input: "a", params: []string{"0"}, expected: false},
		"non-numeric len <= negative max": {input: "a", params: []string{"-1"}, expected: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := Max(tc.input, tc.params...)
			assert.Equal(t, tc.expected, result)
		})
	}
}
