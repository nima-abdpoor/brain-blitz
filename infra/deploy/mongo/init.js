// Re-running this job keeps the existing replica set and data.
try {
  rs.status();
} catch (error) {
  if (error.codeName !== "NotYetInitialized" && error.code !== 94) throw error;
  rs.initiate({
    _id: "rs0",
    members: [
      { _id: 0, host: "bb-mongo1:27017" },
      { _id: 1, host: "bb-mongo2:27017" },
      { _id: 2, host: "bb-mongo3:27017" }
    ]
  });
}
let ready = false;
for (let attempt = 0; attempt < 90; attempt++) {
  if (rs.status().members.some(member => member.state === 1 && member.health === 1)) {
    ready = true;
    break;
  }
  sleep(1000);
}
if (!ready) throw new Error("MongoDB replica set did not elect its primary within 90 seconds");
print("MongoDB replica set is ready");
