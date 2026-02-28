"use client";

import { Suspense } from "react";
import AcceptInvite from "../../../src/screens/public/AcceptInvite";

export default function Page() {
  return (
    <Suspense fallback={null}>
      <AcceptInvite />
    </Suspense>
  );
}
