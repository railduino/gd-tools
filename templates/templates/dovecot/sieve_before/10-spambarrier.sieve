require ["fileinto", "imap4flags"];

if anyof (
  header :contains "X-Spambarrier-Spam" "true",
  header :contains "X-Spam-Flag" "YES"
) {
  addflag "\\Seen";
  fileinto "Junk";
  stop;
}

